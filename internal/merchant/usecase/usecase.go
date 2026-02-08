package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/fekuna/omnipos-pkg/logger"
	"github.com/fekuna/omnipos-user-service/internal/merchant"
	"github.com/fekuna/omnipos-user-service/internal/merchant/dto"
	"github.com/fekuna/omnipos-user-service/internal/model"
	"github.com/fekuna/omnipos-user-service/internal/refreshtoken"
)

var (
	ErrMerchantNotFound   = errors.New("merchant not found")
	ErrInvalidCredentials = errors.New("invalid phone or PIN")
)

// merchantUsecase implements merchant.MerchantUsecase interface
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
) merchant.MerchantUsecase {
	return &merchantUsecase{
		merchantRepo:       merchantRepo,
		refreshTokenRepo:   refreshTokenRepo,
		logger:             log,
		jwtSecretKey:       jwtSecretKey,
		accessTokenExpiry:  accessTokenExpiry,
		refreshTokenExpiry: refreshTokenExpiry,
	}
}

func (uc *merchantUsecase) GetMerchantByPhone(ctx context.Context, phone string) (*model.Merchant, error) {
	merchant, err := uc.merchantRepo.FindOneByAttributes(ctx, &dto.FindOneByAttribute{
		Phone: phone,
	})
	if err != nil {
		return nil, err
	}
	if merchant == nil {
		return nil, ErrMerchantNotFound
	}
	return merchant, nil
}
