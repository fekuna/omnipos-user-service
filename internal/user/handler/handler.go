package handler

import (
	"context"
	"time"

	"github.com/fekuna/omnipos-pkg/audit"
	"github.com/fekuna/omnipos-pkg/logger"
	userv1 "github.com/fekuna/omnipos-proto/proto/user/v1"
	"github.com/fekuna/omnipos-user-service/internal/auth"
	"github.com/fekuna/omnipos-user-service/internal/user/usecase"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type UserHandler struct {
	userv1.UnimplementedUserServiceServer
	uc             usecase.Usecase
	logger         logger.ZapLogger
	auditPublisher *audit.AuditPublisher
}

func NewUserHandler(uc usecase.Usecase, logger logger.ZapLogger, auditPublisher *audit.AuditPublisher) *UserHandler {
	return &UserHandler{
		uc:             uc,
		logger:         logger,
		auditPublisher: auditPublisher,
	}
}

// getRequestMetadata extracts IP and UserAgent from gRPC metadata
func getRequestMetadata(ctx context.Context) (ipAddress, userAgent string) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get("x-forwarded-for"); len(vals) > 0 {
			ipAddress = vals[0]
		}
		if vals := md.Get("user-agent"); len(vals) > 0 {
			userAgent = vals[0]
		}
	}
	return
}

func (h *UserHandler) CreateUser(ctx context.Context, req *userv1.CreateUserRequest) (*userv1.CreateUserResponse, error) {
	startTime := time.Now()
	merchantID := auth.GetMerchantID(ctx)
	userID := auth.GetUserID(ctx)

	if merchantID == "" {
		return nil, status.Error(codes.Unauthenticated, "merchant_id missing")
	}

	user, err := h.uc.CreateUser(ctx, req, merchantID)
	if err != nil {
		h.logger.Error("failed to create user", zap.Error(err))
		// Publish failed audit event
		if h.auditPublisher != nil {
			h.auditPublisher.Publish(ctx, audit.AuditPayload{
				MerchantID:   merchantID,
				UserID:       userID,
				Action:       "user.create",
				EntityType:   "user",
				Result:       "failure",
				ErrorMessage: err.Error(),
				Severity:     "warning",
				DurationMs:   time.Since(startTime).Milliseconds(),
			})
		}
		return nil, status.Error(codes.Internal, "failed to create user")
	}

	// Publish success audit event
	if h.auditPublisher != nil {
		h.auditPublisher.PublishCRUD(ctx, "user.create", "user", user.Id, merchantID, userID, nil, map[string]interface{}{
			"email":    user.Email,
			"username": user.Username,
		})
	}

	return &userv1.CreateUserResponse{User: user}, nil
}

func (h *UserHandler) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	user, err := h.uc.GetUser(ctx, req.Id)
	if err != nil {
		h.logger.Error("failed to get user", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get user")
	}
	if user == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	return &userv1.GetUserResponse{User: user}, nil
}

func (h *UserHandler) ListUsers(ctx context.Context, req *userv1.ListUsersRequest) (*userv1.ListUsersResponse, error) {
	merchantID := auth.GetMerchantID(ctx)
	if merchantID == "" {
		h.logger.Error("merchant_id missing from context")
		return nil, status.Error(codes.Unauthenticated, "merchant_id missing")
	}

	res, err := h.uc.ListUsers(ctx, merchantID, req)
	if err != nil {
		h.logger.Error("failed to list users", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to list users")
	}
	return res, nil
}

func (h *UserHandler) UpdateUser(ctx context.Context, req *userv1.UpdateUserRequest) (*userv1.UpdateUserResponse, error) {
	merchantID := auth.GetMerchantID(ctx)
	userID := auth.GetUserID(ctx)

	user, err := h.uc.UpdateUser(ctx, req)
	if err != nil {
		h.logger.Error("failed to update user", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update user")
	}

	// Publish audit event
	if h.auditPublisher != nil {
		h.auditPublisher.PublishCRUD(ctx, "user.update", "user", req.Id, merchantID, userID, nil, map[string]interface{}{
			"updated_fields": "email,username",
		})
	}

	return &userv1.UpdateUserResponse{User: user}, nil
}

func (h *UserHandler) DeleteUser(ctx context.Context, req *userv1.DeleteUserRequest) (*userv1.DeleteUserResponse, error) {
	merchantID := auth.GetMerchantID(ctx)
	userID := auth.GetUserID(ctx)

	err := h.uc.DeleteUser(ctx, req.Id)
	if err != nil {
		h.logger.Error("failed to delete user", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to delete user")
	}

	// Publish audit event
	if h.auditPublisher != nil {
		h.auditPublisher.PublishCRUD(ctx, "user.delete", "user", req.Id, merchantID, userID, nil, nil)
	}

	return &userv1.DeleteUserResponse{Success: true}, nil
}

func (h *UserHandler) LoginUser(ctx context.Context, req *userv1.LoginUserRequest) (*userv1.LoginUserResponse, error) {
	startTime := time.Now()
	ipAddress, userAgent := getRequestMetadata(ctx)

	// 1. Authenticate User
	user, accessToken, refreshToken, err := h.uc.LoginUser(ctx, req, req.MerchantId)
	if err != nil {
		h.logger.Error("failed to login user", zap.Error(err))

		// Publish failed login audit
		if h.auditPublisher != nil {
			h.auditPublisher.PublishLogin(ctx, req.MerchantId, "", "failure", ipAddress, userAgent, time.Since(startTime).Milliseconds(), map[string]interface{}{
				"username": req.Username,
				"reason":   err.Error(),
			})
		}

		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	// Publish successful login audit
	if h.auditPublisher != nil {
		h.auditPublisher.PublishLogin(ctx, req.MerchantId, user.Id, "success", ipAddress, userAgent, time.Since(startTime).Milliseconds(), map[string]interface{}{
			"username": user.Username,
		})
	}

	return &userv1.LoginUserResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}
