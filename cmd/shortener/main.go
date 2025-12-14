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

	// через этот канал сообщим основному потоку, что соединения закрыты
	idleConnsClosed := make(chan struct{})
	// канал для перенаправления прерываний
	// поскольку нужно отловить всего одно прерывание,
	// ёмкости 1 для канала будет достаточно
	sigint := make(chan os.Signal, 1)
	// регистрируем перенаправление прерываний
	signal.Notify(sigint, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	// запускаем горутину обработки пойманных прерываний
	go func() {
		// читаем из канала прерываний
		// поскольку нужно прочитать только одно прерывание,
		// можно обойтись без цикла
		<-sigint
		logger.Log.Info("shutting down gracefully, press Ctrl+C again to force")

		// получили сигнал os.Interrupt, запускаем процедуру graceful shutdown
		// контекст с таймаутом 5 секунд для завершения обработки текущих запросов
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			// ошибки закрытия Listener
			logger.Log.Error(fmt.Sprintf("HTTP server Shutdown: %v", err))
		}
		// сообщаем основному потоку,
		// что все сетевые соединения обработаны и закрыты
		close(idleConnsClosed)
	}()

	if cfg.Server.EnableHTTPS {
		logger.Log.Info("Starting server with HTTPS", zap.String("address", cfg.Server.Address))

		tlsConfig, err := createTLSConfig()
		if err != nil {
			logger.Log.Fatal(fmt.Errorf("failed to create TLS config: %w", err).Error())
		}
		srv.TLSConfig = tlsConfig

		if err := srv.ListenAndServeTLS("", ""); err != http.ErrServerClosed {
			// ошибки старта или остановки Listener
			logger.Log.Fatal(fmt.Errorf("HTTP server ListenAndServeTLS: %w", err).Error())
		}
	} else {
		logger.Log.Info("Starting server with HTTP", zap.String("address", cfg.Server.Address))
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			// ошибки старта или остановки Listener
			logger.Log.Fatal(fmt.Errorf("HTTP server ListenAndServe: %w", err).Error())
		}
	}

	// ждём завершения процедуры graceful shutdown
	<-idleConnsClosed

	// получили оповещение о завершении
	// здесь освобождаем ресурсы перед выходом
	if db != nil {
		if err := db.Close(); err != nil {
			logger.Log.Error(fmt.Sprintf("Error closing database: %v", err))
		}
	}

	utils.LogErrorWrapper(linksRepository.SaveInStorage())
	logger.Log.Sync()

	logger.Log.Info("Server Shutdown gracefully")
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
