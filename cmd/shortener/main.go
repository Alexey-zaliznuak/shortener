package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"database/sql"
	"encoding/pem"
	"flag"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Alexey-zaliznuak/shortener/internal/config"
	"github.com/Alexey-zaliznuak/shortener/internal/handler"
	"github.com/Alexey-zaliznuak/shortener/internal/handler/audit"
	"github.com/Alexey-zaliznuak/shortener/internal/logger"
	"github.com/Alexey-zaliznuak/shortener/internal/repository/database"
	"github.com/Alexey-zaliznuak/shortener/internal/repository/link"
	"github.com/Alexey-zaliznuak/shortener/internal/service"
	"github.com/Alexey-zaliznuak/shortener/internal/utils"
	"go.uber.org/zap"

	_ "net/http/pprof"
)

const (
	certFile = "cert.pem"
	keyFile  = "key.pem"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	var db *sql.DB

	// Init config
	flagsConfig := config.CreateFLagsInitialConfig()
	flag.Parse()

	cfg, err := config.GetConfig(flagsConfig)

	if err != nil {
		logger.Log.Error(err.Error())
	}

	logger.Initialize(cfg.LoggingLevel)

	logger.Log.Info("Configuration", zap.Any("config", cfg))

	// Init dependencies
	if cfg.DB.DatabaseDSN != "" {
		db, err = database.NewDatabaseConnectionPool(cfg)
		if err != nil {
			logger.Log.Fatal(err.Error())
		}
	}

	linksRepository, err := link.NewLinksRepository(context.Background(), cfg, db)
	if err != nil {
		logger.Log.Fatal(err.Error())
	}

	if err := linksRepository.LoadStoredData(); err != nil {
		logger.Log.Fatal(err.Error())
	}

	linksService := service.NewLinksService(linksRepository, cfg)

	auditor := audit.NewAuditorShortURLOperationManager()

	if cfg.Audit.AuditFile != "" {
		auditor.UseAuditor(&audit.AuditShortURLOperationFile{FilePath: cfg.Audit.AuditFile})
	}

	if cfg.Audit.AuditURL != "" {
		auditor.UseAuditor(&audit.AuditShortURLOperationHTTP{URL: cfg.Audit.AuditURL})
	}

	router := handler.NewRouter()
	authService := service.NewAuthService(cfg)
	handler.RegisterLinksRoutes(router, linksService, authService, auditor, db)
	handler.RegisterAppHandlerRoutes(router, db)

	// Server process
	srv := &http.Server{Addr: cfg.Server.Address, Handler: router}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	go func() {
		<-ctx.Done()
		logger.Log.Info("shutting down gracefully, press Ctrl+C again to force")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Log.Error(fmt.Sprintf("HTTP server Shutdown: %v", err))
		}
	}()

	if cfg.Server.EnableHTTPS {
		logger.Log.Info("Starting server with HTTPS", zap.String("address", cfg.Server.Address))

		tlsConfig, err := createTLSConfig()
		if err != nil {
			logger.Log.Fatal(fmt.Errorf("failed to create TLS config: %w", err).Error())
		}
		srv.TLSConfig = tlsConfig

		if err := srv.ListenAndServeTLS("", ""); err != http.ErrServerClosed {
			logger.Log.Fatal(fmt.Errorf("HTTP server ListenAndServeTLS: %w", err).Error())
		}
	} else {
		logger.Log.Info("Starting server with HTTP", zap.String("address", cfg.Server.Address))
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			logger.Log.Fatal(fmt.Errorf("HTTP server ListenAndServe: %w", err).Error())
		}
	}

	if db != nil {
		if err := db.Close(); err != nil {
			logger.Log.Error(fmt.Sprintf("Error closing database: %v", err))
		}
	}

	utils.LogErrorWrapper(linksRepository.SaveInStorage())
	logger.Log.Sync()

	logger.Log.Info("Server Shutdown gracefully")
}

// createTLSConfig создает конфигурацию TLS с сертификатом.
func createTLSConfig() (*tls.Config, error) {
	cert, err := getOrCreateCert()
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
	}, nil
}

func getOrCreateCert() (tls.Certificate, error) {
	if fileExists(certFile) && fileExists(keyFile) {
		logger.Log.Info("Loading existing certificate", zap.String("cert", certFile), zap.String("key", keyFile))
		return tls.LoadX509KeyPair(certFile, keyFile)
	}

	logger.Log.Info("Certificate files not found, generating new certificate")

	certPEM, keyPEM, err := generateCert()
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to generate certificate: %w", err)
	}

	// Сохраняем сертификат и ключ в файлы
	if err := os.WriteFile(certFile, certPEM, 0644); err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to save certificate: %w", err)
	}

	if err := os.WriteFile(keyFile, keyPEM, 0600); err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to save private key: %w", err)
	}

	logger.Log.Info("Certificate saved", zap.String("cert", certFile), zap.String("key", keyFile))

	return tls.X509KeyPair(certPEM, keyPEM)
}

// generateCert генерирует самоподписанный сертификат и приватный ключ.
func generateCert() (certPEM []byte, keyPEM []byte, err error) {
	// Генерируем приватный ключ RSA
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Создаем шаблон сертификата
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Shortener"},
			CommonName:   "localhost",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // Сертификат действителен 1 год
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost"},
	}

	// Создаем самоподписанный сертификат
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	// Кодируем сертификат в PEM
	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	// Кодируем приватный ключ в PEM
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})

	return certPEM, keyPEM, nil
}

// fileExists проверяет существование файла.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
