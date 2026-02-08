package config

import (
	"os"
	"strings"
	"time"
)

type Config struct {
	Server   ServerConfig
	GRPC     GRPCConfig
	Postgres PostgresConfig
	Logger   LoggerConfig
	JWT      JWTConfig
	Kafka    KafkaConfig
}

type ServerConfig struct {
	AppName    string
	AppEnv     string
	PrivateKey string
}

type GRPCConfig struct {
	Port string
}

type PostgresConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	SSLMode         string // disable | require | verify-full
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

type LoggerConfig struct {
	Level             string
	Encoding          string
	DisableCaller     bool
	DisableStacktrace bool
}

type JWTConfig struct {
	SecretKey          string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
}

type KafkaConfig struct {
	Brokers []string
}

func LoadEnv() *Config {
	return &Config{
		Server: ServerConfig{
			AppName:    getEnv("APP_NAME", "omnipos-user-service"),
			AppEnv:     getEnv("APP_ENV", "dev"),
			PrivateKey: getEnvRequired("PRIVATE_KEY"),
		},
		GRPC: GRPCConfig{
			Port: getEnv("GRPC_PORT", ":8080"),
		},
		Postgres: PostgresConfig{
			Host:            getEnvRequired("POSTGRES_HOST"),
			Port:            getEnvRequired("POSTGRES_PORT"),
			User:            getEnvRequired("POSTGRES_USER"),
			Password:        getEnvRequired("POSTGRES_PASSWORD"),
			DBName:          getEnvRequired("POSTGRES_DB_NAME"),
			SSLMode:         getEnv("disable", "disable"),
			MaxOpenConns:    getEnvInt("POSTGRES_MAX_OPEN_CONNS", 10),
			MaxIdleConns:    getEnvInt("POSTGRES_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvDuration("POSTGRES_CONN_MAX_LIFETIME", 5*time.Minute),
			ConnMaxIdleTime: getEnvDuration("POSTGRES_CONN_MAX_IDLE_TIME", 1*time.Minute),
		},
		Logger: LoggerConfig{
			Level:             getEnv("LOG_LEVEL", "info"),
			Encoding:          getEnv("LOG_ENCODING", "json"),
			DisableCaller:     getBoolEnv("LOG_DISABLE_CALLER", false),
			DisableStacktrace: getBoolEnv("LOG_DISABLE_STACKTRACE", false),
		},
		JWT: JWTConfig{
			SecretKey:          getEnvRequired("JWT_SECRET_KEY"),
			AccessTokenExpiry:  getEnvDuration("JWT_ACCESS_TOKEN_EXPIRY", 15*time.Minute),
			RefreshTokenExpiry: getEnvDuration("JWT_REFRESH_TOKEN_EXPIRY", 168*time.Hour),
		},
		Kafka: KafkaConfig{
			Brokers: getKafkaBrokers(),
		},
	}
}

func getKafkaBrokers() []string {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		return []string{}
	}
	return strings.Split(brokers, ",")
}
