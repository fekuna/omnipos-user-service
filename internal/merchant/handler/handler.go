package handler

import (
	"context"
	"errors"

	"github.com/fekuna/omnipos-pkg/logger"
	userv1 "github.com/fekuna/omnipos-proto/gen/go/omnipos/user/v1"
	"github.com/fekuna/omnipos-user-service/internal/auth"
	"github.com/fekuna/omnipos-user-service/internal/merchant"
	"github.com/fekuna/omnipos-user-service/internal/merchant/usecase"
	useruc "github.com/fekuna/omnipos-user-service/internal/user/usecase"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MerchantHandler struct {
	userv1.UnimplementedMerchantServiceServer

	merchantUsecase merchant.MerchantUsecase
	userUsecase     useruc.Usecase
	logger          logger.ZapLogger
}

// NewMerchantHandler creates a new merchant gRPC handler
func NewMerchantHandler(merchantUsecase merchant.MerchantUsecase, userUsecase useruc.Usecase, log logger.ZapLogger) *MerchantHandler {
	return &MerchantHandler{
		merchantUsecase: merchantUsecase,
		userUsecase:     userUsecase,
		logger:          log,
	}
}

func (h *MerchantHandler) LoginMerchant(ctx context.Context, req *userv1.LoginMerchantRequest) (*userv1.LoginMerchantResponse, error) {
	h.logger.Info("processing login request", zap.String("phone", req.Phone))

	// Call use case
	accessToken, refreshToken, err := h.merchantUsecase.Login(ctx, req.Phone, req.Pin)
	if err != nil {
		h.logger.Error("login failed", zap.Error(err))

		// Map errors to gRPC status codes
		if errors.Is(err, usecase.ErrMerchantNotFound) {
			return nil, status.Error(codes.NotFound, "merchant not found")
		}
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "invalid phone or PIN")
		}

		return nil, status.Error(codes.Internal, "internal server error")
	}

	// Check if user management is enabled
	// Since we don't have the merchant object here (Login returns tokens only),
	// we technically should modify Usecase.Login to return Merchant object OR fetch it here.
	// But h.merchantUsecase.Login takes phone/pin and return tokens.
	// To fit the "Instant Accuracy" plan, we need the Merchant object to check FeatureFlags.

	// Let's fetch the merchant by phone to get features.
	// Note: Login already validates credentials, so fetching by phone is safe here IF login succeeds.
	merchantObj, err := h.merchantUsecase.GetMerchantByPhone(ctx, req.Phone)
	var availableUsers []*userv1.UserInfo
	var userManagementEnabled bool

	if err == nil {
		userManagementEnabled = merchantObj.FeatureFlags.UserManagement

		if userManagementEnabled {
			// Fetch users
			// We iterate pages or just fetch first page? Ideally fetch all active users (lite version).
			// userUsecase.ListUsers requires pagination. Let's ask for 100 users for now.
			listResp, err := h.userUsecase.ListUsers(ctx, merchantObj.ID, &userv1.ListUsersRequest{
				Page:     1,
				PageSize: 100, // Reasonable limit for login screen
			})
			if err == nil {
				for _, u := range listResp.Users {
					if u.Status == "active" {
						availableUsers = append(availableUsers, &userv1.UserInfo{
							Id:       u.Id,
							Username: u.Username,
							FullName: u.FullName,
							RoleName: u.Role.Name, // Assuming Role is populated in User
						})
					}
				}
			} else {
				h.logger.Error("failed to list users for merchant login", zap.Error(err))
			}
		}
	} else {
		h.logger.Error("failed to fetch merchant details after login", zap.Error(err))
	}

	h.logger.Info("login successful", zap.String("phone", req.Phone))

	return &userv1.LoginMerchantResponse{
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		UserManagementEnabled: userManagementEnabled,
		AvailableUsers:        availableUsers,
	}, nil
}

// LogoutMerchant handles the merchant logout gRPC request (single device)
// Note: This signature will be updated once proto is generated
func (h *MerchantHandler) LogoutMerchant(ctx context.Context, req interface{}) (interface{}, error) {
	type LogoutRequest struct {
		RefreshToken string
	}

	type LogoutResponse struct {
		Success bool
		Message string
	}

	logoutReq, ok := req.(*LogoutRequest)
	if !ok {
		h.logger.Error("invalid request type")
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	h.logger.Info("processing logout request")

	// Call use case
	err := h.merchantUsecase.Logout(ctx, logoutReq.RefreshToken)
	if err != nil {
		h.logger.Error("logout failed", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to logout")
	}

	h.logger.Info("logout successful")

	return &LogoutResponse{
		Success: true,
		Message: "logged out successfully",
	}, nil
}

// LogoutAllDevices handles logout from all devices for a merchant
// Note: This signature will be updated once proto is generated
func (h *MerchantHandler) LogoutAllDevices(ctx context.Context, req interface{}) (interface{}, error) {
	type LogoutAllRequest struct {
		MerchantID string
	}

	type LogoutAllResponse struct {
		Success bool
		Message string
	}

	logoutAllReq, ok := req.(*LogoutAllRequest)
	if !ok {
		h.logger.Error("invalid request type")
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	h.logger.Info("processing logout all devices request", zap.String("merchant_id", logoutAllReq.MerchantID))

	// Call use case
	err := h.merchantUsecase.LogoutAllDevices(ctx, logoutAllReq.MerchantID)
	if err != nil {
		h.logger.Error("logout all devices failed", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to logout from all devices")
	}

	h.logger.Info("logout all devices successful", zap.String("merchant_id", logoutAllReq.MerchantID))

	return &LogoutAllResponse{
		Success: true,
		Message: "logged out from all devices successfully",
	}, nil
}

// RefreshToken handles refresh token rotation to get a new access token
func (h *MerchantHandler) RefreshToken(ctx context.Context, req *userv1.RefreshTokenRequest) (*userv1.RefreshTokenResponse, error) {
	h.logger.Info("processing refresh token request")

	// Call use case
	accessToken, refreshToken, err := h.merchantUsecase.RefreshAccessToken(ctx, req.RefreshToken)
	if err != nil {
		h.logger.Error("token refresh failed", zap.Error(err))

		// Map errors to gRPC status codes
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "invalid or revoked refresh token")
		}

		return nil, status.Error(codes.Internal, "failed to refresh token")
	}

	h.logger.Info("token refresh successful")

	return &userv1.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// GetCurrentMerchant handles the current merchant retrieval request
// The merchant is identified by the JWT token (via context)
func (h *MerchantHandler) GetCurrentMerchant(ctx context.Context, req *emptypb.Empty) (*userv1.GetCurrentMerchantResponse, error) {
	// Extract user context from context (added by auth interceptor)
	userCtx := auth.MustGetUserContext(ctx)

	h.logger.Info("processing get current merchant request", zap.String("merchant_id", userCtx.MerchantID))

	// Call use case to get merchant details using ID from token
	merchant, err := h.merchantUsecase.GetMerchantDetail(ctx, userCtx.MerchantID)
	if err != nil {
		h.logger.Error("failed to get merchant detail", zap.Error(err))

		// Map errors to gRPC status codes
		if errors.Is(err, usecase.ErrMerchantNotFound) {
			return nil, status.Error(codes.NotFound, "merchant not found")
		}

		return nil, status.Error(codes.Internal, "internal server error")
	}

	h.logger.Info("merchant detail retrieved successfully", zap.String("merchant_id", userCtx.MerchantID))

	// Map domain model to proto response
	return &userv1.GetCurrentMerchantResponse{
		Id:                    merchant.ID,
		Name:                  merchant.Name,
		Phone:                 merchant.Phone,
		Timezone:              merchant.Timezone,
		CreatedAt:             timestamppb.New(merchant.CreatedAt),
		UpdatedAt:             timestamppb.New(merchant.UpdatedAt),
		UserManagementEnabled: merchant.FeatureFlags.UserManagement,
	}, nil
}
