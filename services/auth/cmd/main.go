package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"auth/internal/app"
	"auth/internal/config"
	"auth/internal/observability"
)

func main() {
	log.Println("starting auth service")

	cfg := config.MustLoad()
	log.Println("configuration loaded")

	logger := observability.NewLogger(cfg)
	logger.Info("logger initialized")

	observability.Register("consul:8500", "auth-service", 50053)
	logger.Info("service registration completed")
	// logger := observability.Register(cfg)
	// tracer := observability.NewTracer(cfg)

	container := app.BuildContainer(cfg, logger)
	logger.Info("application container built")

	application := app.New(container)
	logger.Info("starting application lifecycle")

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	if err := application.Run(ctx); err != nil {
		logger.Error("application exited with error", "error", err)
		os.Exit(1)
	}

	logger.Info("application exited normally")
}
