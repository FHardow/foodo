package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fhardow/bread-order/internal/config"
	"github.com/fhardow/bread-order/internal/domain/order"
	"github.com/fhardow/bread-order/internal/domain/product"
	"github.com/fhardow/bread-order/internal/domain/user"
	apphttp "github.com/fhardow/bread-order/internal/infra/http"
	"github.com/fhardow/bread-order/internal/infra/http/handler"
	"github.com/fhardow/bread-order/internal/infra/postgres"
	"github.com/fhardow/bread-order/internal/infra/telegram"
	"github.com/fhardow/bread-order/pkg/logger"
)

const uploadsDir = "./uploads"

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	log := logger.New(cfg.Env)

	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		log.Error("failed to create uploads directory", "err", err)
		os.Exit(1)
	}

	db, err := postgres.Connect(cfg.DSN, cfg.Env)
	if err != nil {
		log.Error("failed to connect to database", "err", err)
		os.Exit(1)
	}

	if err := postgres.Migrate(db); err != nil {
		log.Error("failed to run migrations", "err", err)
		os.Exit(1)
	}

	// Repositories
	userRepo    := postgres.NewUserRepo(db)
	productRepo := postgres.NewProductRepo(db)
	orderRepo   := postgres.NewOrderRepo(db)

	// Telegram notifier (optional — only active when both env vars are set)
	var orderNotifier order.Notifier
	if cfg.TelegramBotToken != "" && cfg.TelegramChatID != "" {
		orderNotifier = telegram.New(cfg.TelegramBotToken, cfg.TelegramChatID)
		log.Info("telegram notifications enabled")
	}

	// Domain services
	userSvc    := user.NewService(userRepo)
	productSvc := product.NewService(productRepo)
	orderSvc   := order.NewService(orderRepo, productRepo, orderNotifier)

	// HTTP handlers
	userHandler    := handler.NewUserHandler(userSvc)
	productHandler := handler.NewProductHandler(productSvc, uploadsDir)
	orderHandler   := handler.NewOrderHandler(orderSvc)

	router := apphttp.NewRouter(userHandler, productHandler, orderHandler, cfg.KeycloakURL, cfg.KeycloakRealm, uploadsDir)
	srv    := apphttp.NewServer(cfg.Port, router, log)

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("shutdown error", "err", err)
	}
}
