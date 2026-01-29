package usecase

import (
	"errors"
	"time"

	"github.com/fekuna/omnipos-pkg/logger"
	"github.com/fekuna/omnipos-user-service/internal/merchant"
	"github.com/fekuna/omnipos-user-service/internal/refreshtoken"
)

var (
	ErrMerchantNotFound   = errors.New("merchant not found")
	ErrInvalidCredentials = errors.New("invalid phone or PIN")
)

// MerchantUsecase defines the business logic interface for merchant operations
type MerchantUsecase interface {
	Login(phone, pin string) (accessToken, refreshToken string, err error)
	Logout(refreshToken string) error
	LogoutAllDevices(merchantID string) error
	RefreshAccessToken(refreshToken string) (accessToken string, err error)
}

// merchantUsecase implements MerchantUsecase interface
type merchantUsecase struct {
	merchantRepo       merchant.PGRepository
	refreshTokenRepo   refreshtoken.Repository
	logger             logger.ZapLogger
	jwtSecretKey       string
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
}

// NewMerchantUsecase creates a new merchant usecase instance
func NewMerchantUsecase(
	merchantRepo merchant.PGRepository,
	refreshTokenRepo refreshtoken.Repository,
	log logger.ZapLogger,
	jwtSecretKey string,
	accessTokenExpiry time.Duration,
	refreshTokenExpiry time.Duration,
) MerchantUsecase {
	return &merchantUsecase{
		merchantRepo:       merchantRepo,
		refreshTokenRepo:   refreshTokenRepo,
		logger:             log,
		jwtSecretKey:       jwtSecretKey,
		accessTokenExpiry:  accessTokenExpiry,
		refreshTokenExpiry: refreshTokenExpiry,
	}
}
