package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"link/internal/config"
	"link/internal/handler"
	"link/internal/middleware"
	"link/internal/pkg/cardtoken"
	"link/internal/pkg/token"
	"link/internal/repository/postgres"
	"link/internal/service"
	"link/internal/transport"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	logLevel := slog.LevelInfo
	if cfg.LogLevel == "debug" {
		logLevel = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	tokenMgr := token.NewManager(cfg.JWTSecret, cfg.JWTExpiry)
	cardTokenGen := cardtoken.NewGenerator(cfg.CardTokenSecret)

	userRepo := postgres.NewUserRepository(pool)
	cardRepo := postgres.NewCardRepository(pool)
	sessionRepo := postgres.NewSessionRepository(pool)
	friendRepo := postgres.NewFriendshipRepository(pool)
	convRepo := postgres.NewConversationRepository(pool)
	msgRepo := postgres.NewMessageRepository(pool)

	userSvc := service.NewUserService(userRepo)
	cardSvc := service.NewCardService(cardRepo, sessionRepo, cardTokenGen)
	authSvc := service.NewAuthService(userRepo, cardRepo, sessionRepo, tokenMgr, cardTokenGen)
	friendSvc := service.NewFriendshipService(friendRepo, userRepo)
	convSvc := service.NewConversationService(convRepo)
	msgSvc := service.NewMessageService(msgRepo, convRepo)

	hub := transport.NewHub()
	go hub.Run()

	transportHandler := transport.NewHandler(hub, msgSvc, convSvc)

	authHandler := handler.NewAuthHandler(authSvc, cardSvc)
	userHandler := handler.NewUserHandler(userSvc, cardSvc)
	friendHandler := handler.NewFriendHandler(friendSvc)
	convHandler := handler.NewConversationHandler(convSvc, msgSvc, hub)
	adminHandler := handler.NewAdminHandler(cardTokenGen, cfg.AdminPassword, cfg.BaseURL)

	handlers := &handler.Handlers{
		Auth:   authHandler,
		User:   userHandler,
		Friend: friendHandler,
		Conv:   convHandler,
		Admin:  adminHandler,
	}

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return handler.Error(c, err)
		},
	})

	app.Use(recover.New())
	app.Use(middleware.Logger())
	app.Use(middleware.SecurityHeaders())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSOrigins,
		AllowMethods:     "GET,POST,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Admin-Password",
		AllowCredentials: true,
	}))

	authMw := middleware.Auth(tokenMgr)
	handler.Setup(app, handlers, authMw)

	wsServer := transport.NewServer(hub, transportHandler, tokenMgr)
	wsServer.SetupRoutes(app)

	go func() {
		var listenErr error
		if strings.HasSuffix(cfg.ServerAddr, "443") && cfg.TLSCert != "" && cfg.TLSKey != "" {
			slog.Info("starting HTTPS server", "addr", cfg.ServerAddr)
			listenErr = app.ListenTLS(cfg.ServerAddr, cfg.TLSCert, cfg.TLSKey)
		} else {
			slog.Info("starting HTTP server", "addr", cfg.ServerAddr)
			listenErr = app.Listen(cfg.ServerAddr)
		}
		if listenErr != nil {
			log.Fatalf("server error: %v", listenErr)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server...")
	if err := app.ShutdownWithTimeout(10 * time.Second); err != nil {
		slog.Error("server shutdown error", "err", err)
	}
}
