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
	defer logger.Log.Sync()

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
	defer func() { utils.LogErrorWrapper(linksRepository.SaveInStorage()) }()

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
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{Addr: cfg.Server.Address, Handler: router}

	go func() {
		if cfg.Server.EnableHTTPS {
			logger.Log.Info("Starting server with HTTPS", zap.String("address", cfg.Server.Address))

			tlsConfig, err := createTLSConfig()
			if err != nil {
				logger.Log.Fatal(fmt.Errorf("failed to create TLS config: %w", err).Error())
			}
			srv.TLSConfig = tlsConfig

			if err := srv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
				logger.Log.Fatal(fmt.Errorf("listen TLS: %w", err).Error())
			}
		} else {
			logger.Log.Info("Starting server with HTTP", zap.String("address", cfg.Server.Address))
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Log.Fatal(fmt.Errorf("listen: %w", err).Error())
			}
		}
	}()

	// go func() {
	// 	logger.Log.Info("pprof listening on :9090")
	// 	http.ListenAndServe(":9090", nil)
	// }()

	// Listen for the interrupt signal.
	<-ctx.Done()

	stop()

	logger.Log.Info("shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Fatal(fmt.Errorf("server forced to shutdown: %w", err).Error())
	}

	logger.Log.Info("Server exited")
}

// createTLSConfig создает конфигурацию TLS с самоподписанным сертификатом.
func createTLSConfig() (*tls.Config, error) {
	// Генерируем приватный ключ RSA
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
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
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	// Кодируем сертификат в PEM
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	// Кодируем приватный ключ в PEM
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})

	// Загружаем сертификат и ключ
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to load key pair: %w", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
	}, nil
}
