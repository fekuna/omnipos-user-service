package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/fekuna/omnipos-pkg/database/postgres"
	"github.com/fekuna/omnipos-pkg/logger"
	userv1 "github.com/fekuna/omnipos-proto/proto/user/v1"
	"github.com/fekuna/omnipos-user-service/config"
	"github.com/fekuna/omnipos-user-service/internal/merchant/handler"
	merchantRepo "github.com/fekuna/omnipos-user-service/internal/merchant/repository"
	"github.com/fekuna/omnipos-user-service/internal/merchant/usecase"
	refreshTokenRepo "github.com/fekuna/omnipos-user-service/internal/refreshtoken/repository"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	_ = godotenv.Load()

	cfg := config.LoadEnv()

	loggerCfg := logger.ZapLoggerConfig{
		IsDevelopment:     cfg.Server.AppEnv == "dev",
		Level:             cfg.Logger.Level,
		Encoding:          cfg.Logger.Encoding,
		DisableCaller:     cfg.Logger.DisableCaller,
		DisableStacktrace: cfg.Logger.DisableStacktrace,
	}

	log := logger.NewZapLogger(&loggerCfg)
	defer log.Sync()

	log.Info("Logger initialized")

	pgConfig := postgres.Config{
		Host:            cfg.Postgres.Host,
		Port:            cfg.Postgres.Port,
		User:            cfg.Postgres.User,
		Password:        cfg.Postgres.Password,
		DBName:          cfg.Postgres.DBName,
		SSLMode:         cfg.Postgres.SSLMode,
		MaxOpenConns:    cfg.Postgres.MaxOpenConns,
		MaxIdleConns:    cfg.Postgres.MaxIdleConns,
		ConnMaxLifetime: cfg.Postgres.ConnMaxLifetime,
		ConnMaxIdleTime: cfg.Postgres.ConnMaxIdleTime,
	}

	db, err := postgres.NewPostgres(&pgConfig)
	if err != nil {
		log.Fatal("failed to connect to database", zap.Error(err))
	}

	log.Info("Postgres database connected")

	// Initialize repositories
	merchantRepository := merchantRepo.NewPGRepository(db)
	refreshTokenRepository := refreshTokenRepo.NewPGRepository(db)

	log.Info("Repositories initialized")

	// Initialize use cases
	merchantUsecase := usecase.NewMerchantUsecase(
		merchantRepository,
		refreshTokenRepository,
		log,
		cfg.JWT.SecretKey,
		cfg.JWT.AccessTokenExpiry,
		cfg.JWT.RefreshTokenExpiry,
	)

	log.Info("Use cases initialized")

	// Initialize handlers
	merchantHandler := handler.NewMerchantHandler(merchantUsecase, log)

	log.Info("Handlers initialized")

	// Create gRPC server
	grpcServer := grpc.NewServer()
	userv1.RegisterMerchantServiceServer(grpcServer, merchantHandler)
	reflection.Register(grpcServer)

	log.Info("gRPC server configured")

	lis, err := net.Listen("tcp", cfg.GRPC.Port)
	if err != nil {
		log.Fatal("failed to listen", zap.Error(err))
	}

	log.Info("Server started", zap.String("port", cfg.GRPC.Port))

	// graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Info("shutting down grpc server")
		grpcServer.GracefulStop()
	}()

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("failed to serve", zap.Error(err))
	}
}
