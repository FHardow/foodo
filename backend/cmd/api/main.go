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

	"github.com/fhardow/foodo/internal/config"
	"github.com/fhardow/foodo/internal/domain/order"
	"github.com/fhardow/foodo/internal/domain/product"
	"github.com/fhardow/foodo/internal/domain/push"
	"github.com/fhardow/foodo/internal/domain/user"
	apphttp "github.com/fhardow/foodo/internal/infra/http"
	"github.com/fhardow/foodo/internal/infra/http/handler"
	"github.com/fhardow/foodo/internal/infra/postgres"
	"github.com/fhardow/foodo/internal/infra/telegram"
	"github.com/fhardow/foodo/internal/infra/webpush"
	"github.com/fhardow/foodo/pkg/logger"
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

	// Repositories
	userRepo    := postgres.NewUserRepo(db)
	productRepo := postgres.NewProductRepo(db)
	orderRepo   := postgres.NewOrderRepo(db)
	pushRepo, err := postgres.NewPushSubscriptionRepo(db, cfg.PushEncryptionKey)
	if err != nil {
		log.Error("failed to init push subscription repo", "err", err)
		os.Exit(1)
	}

	// Domain services
	userSvc    := user.NewService(userRepo)
	productSvc := product.NewService(productRepo)
	orderSvc   := order.NewService(orderRepo, productRepo, userRepo)
	pushSvc    := push.NewService(pushRepo)
	if cfg.TelegramBotToken != "" && cfg.TelegramChatID != "" {
		orderSvc.WithNotifier(telegram.NewNotifier(cfg.TelegramBotToken, cfg.TelegramChatID))
		log.Info("telegram order notifications enabled")
	}
	// VAPIDSubject is also required — push services reject an empty "mailto:" JWT claim
	if cfg.VAPIDPublicKey != "" && cfg.VAPIDPrivateKey != "" && cfg.VAPIDSubject != "" {
		orderSvc.WithCustomerNotifier(webpush.NewNotifier(pushRepo, cfg.VAPIDPublicKey, cfg.VAPIDPrivateKey, cfg.VAPIDSubject))
		log.Info("push order notifications enabled")
	}

	// HTTP handlers
	userHandler    := handler.NewUserHandler(userSvc)
	productHandler := handler.NewProductHandler(productSvc, uploadsDir)
	orderHandler   := handler.NewOrderHandler(orderSvc)
	pushHandler    := handler.NewPushHandler(pushSvc)

	router := apphttp.NewRouter(userHandler, productHandler, orderHandler, pushHandler, userSvc, cfg.KeycloakURL, cfg.KeycloakRealm, uploadsDir, cfg.CORSOrigin)
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
