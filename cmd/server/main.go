package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/uncle3dev/velotrax-gateway-go/internal/config"
	"github.com/uncle3dev/velotrax-gateway-go/internal/grpc/client"
	"github.com/uncle3dev/velotrax-gateway-go/internal/router"
	"go.uber.org/zap/zapcore"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logger, err := initLogger(cfg.LogLevel)
	fmt.Printf("DEBUG: Loaded Config: %+v\n", cfg)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting velotrax-gateway-go",
		zap.String("env", cfg.AppEnv),
		zap.Int("port", cfg.AppPort),
	)

	// Initialize gRPC clients
	authClient, err := client.NewAuthClient(cfg.GRPCAuthAddr)
	if err != nil {
		logger.Fatal("Failed to connect to auth service", zap.Error(err))
	}
	defer authClient.Close()

	logger.Info("Connected to auth service", zap.String("addr", cfg.GRPCAuthAddr))

	// Set Gin mode
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Create Gin engine
	engine := gin.New()

	// Setup routes
	router.Setup(engine, authClient, logger, cfg.JWTSecret)

	// HTTP server
	srv := &http.Server{
		Addr:           fmt.Sprintf(":%d", cfg.AppPort),
		Handler:        engine,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Server listening", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server error", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server shutdown error", zap.Error(err))
	}

	logger.Info("Server stopped")
}

func initLogger(level string) (*zap.Logger, error) {
	var cfg zap.Config
	if level == "debug" {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
	}
	cfg.Level = zap.NewAtomicLevelAt(parseLogLevel(level))
	return cfg.Build()
}

func parseLogLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	case "warn":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	default:
		return zap.InfoLevel
	}
}
