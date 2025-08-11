package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/Alexey-zaliznuak/shortener/internal/config"
	"github.com/Alexey-zaliznuak/shortener/internal/handler"
	"github.com/Alexey-zaliznuak/shortener/internal/logger"
	"github.com/Alexey-zaliznuak/shortener/internal/repository"
	"github.com/Alexey-zaliznuak/shortener/internal/service"
	"go.uber.org/zap"
)

func main() {
	flagsConfig := config.CreateFLagsInitialConfig()
	flag.Parse()

	cfg, err := config.GetConfig(flagsConfig)

	if err != nil {
		logger.Log.Error(err.Error())
	}

	logger.Initialize(cfg.LoggingLevel)
	defer logger.Log.Sync()

	logger.Log.Info("Configuration", zap.Any("config", cfg))

	linksRepository := repository.NewLinksRepository(cfg)
	defer linksRepository.SaveInStorage()

	if err := linksRepository.LoadStoredData(); err != nil {
		logger.Log.Error(err.Error())
	}

	linksService := service.NewLinksService(linksRepository, cfg)

	router := handler.NewRouter()
	handler.RegisterLinksRoutes(router, linksService)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{Addr: cfg.Server.Address, Handler: router}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal(fmt.Errorf("listen: %w", err).Error())
		}
	}()

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
