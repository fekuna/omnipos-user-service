package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/fekuna/omnipos-pkg/audit"
	"github.com/fekuna/omnipos-pkg/database/postgres"
	"github.com/fekuna/omnipos-pkg/logger"
	userv1 "github.com/fekuna/omnipos-proto/gen/go/omnipos/user/v1"
	"github.com/fekuna/omnipos-user-service/config"
	"github.com/fekuna/omnipos-user-service/internal/merchant/handler"
	merchantRepo "github.com/fekuna/omnipos-user-service/internal/merchant/repository"
	"github.com/fekuna/omnipos-user-service/internal/merchant/usecase"
	"github.com/fekuna/omnipos-user-service/internal/middleware"
	refreshTokenRepo "github.com/fekuna/omnipos-user-service/internal/refreshtoken/repository"
	roleHandler "github.com/fekuna/omnipos-user-service/internal/role/handler"
	roleRepo "github.com/fekuna/omnipos-user-service/internal/role/repository"
	roleUC "github.com/fekuna/omnipos-user-service/internal/role/usecase"
	userHandler "github.com/fekuna/omnipos-user-service/internal/user/handler"
	userRepo "github.com/fekuna/omnipos-user-service/internal/user/repository"
	userUC "github.com/fekuna/omnipos-user-service/internal/user/usecase"
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
	roleRepository := roleRepo.NewPostgresRepository(db)
	userRepository := userRepo.NewPostgresUserRepository(db)

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
	roleUsecase := roleUC.NewRoleUsecase(roleRepository)
	userUsecase := userUC.NewUserUsecase(
		userRepository,
		merchantUsecase,
		cfg.JWT.SecretKey,
		cfg.JWT.AccessTokenExpiry,
		cfg.JWT.RefreshTokenExpiry,
	)

	log.Info("Use cases initialized")

	// Initialize audit publisher (optional - only if Kafka is configured)
	var auditPublisher *audit.AuditPublisher
	if len(cfg.Kafka.Brokers) > 0 && cfg.Kafka.Brokers[0] != "" {
		auditPublisher = audit.NewAuditPublisher(cfg.Kafka.Brokers, "user-service")
		defer auditPublisher.Close()
		log.Info("Audit publisher initialized", zap.Strings("brokers", cfg.Kafka.Brokers))
	} else {
		log.Warn("Kafka not configured, audit publishing disabled")
	}

	// Initialize handlers
	merchantHandler := handler.NewMerchantHandler(merchantUsecase, userUsecase, log)
	roleHandler := roleHandler.NewRoleHandler(roleUsecase, log)
	userHandler := userHandler.NewUserHandler(userUsecase, log, auditPublisher)

	log.Info("Handlers initialized")

	// Initialize auth context interceptor
	authContextInterceptor := middleware.NewAuthContextInterceptor(log)
	log.Info("Auth context interceptor initialized")

	// Create gRPC server with interceptor
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(authContextInterceptor.Unary()),
	)
	userv1.RegisterMerchantServiceServer(grpcServer, merchantHandler)
	userv1.RegisterRoleServiceServer(grpcServer, roleHandler)
	userv1.RegisterUserServiceServer(grpcServer, userHandler)
	reflection.Register(grpcServer)

	log.Info("gRPC server configured with auth context interceptor")

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
