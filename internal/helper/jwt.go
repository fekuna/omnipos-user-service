package helper

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

// JWTClaims represents the claims stored in the JWT token
type JWTClaims struct {
	MerchantID string `json:"merchant_id"`
	jwt.RegisteredClaims
}

// JWTHelper handles JWT token operations
type JWTHelper struct {
	secretKey          string
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
}

// NewJWTHelper creates a new JWT helper instance
func NewJWTHelper(secretKey string, accessTokenExpiry, refreshTokenExpiry time.Duration) *JWTHelper {
	return &JWTHelper{
		secretKey:          secretKey,
		accessTokenExpiry:  accessTokenExpiry,
		refreshTokenExpiry: refreshTokenExpiry,
	}
}

// GenerateAccessToken generates a short-lived access token for the merchant
func (h *JWTHelper) GenerateAccessToken(merchantID string) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		MerchantID: merchantID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(h.accessTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.secretKey))
}

// GenerateRefreshToken generates a long-lived refresh token for the merchant
func (h *JWTHelper) GenerateRefreshToken(merchantID string) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		MerchantID: merchantID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(h.refreshTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.secretKey))
}

// ValidateToken validates a JWT token and returns the claims
func (h *JWTHelper) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(h.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Check if token is expired
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, ErrExpiredToken
	}

	return claims, nil
}

// ExtractMerchantID extracts the merchant ID from a valid token
func (h *JWTHelper) ExtractMerchantID(tokenString string) (string, error) {
	claims, err := h.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}
	return claims.MerchantID, nil
}

// GetRefreshTokenExpiry returns the expiry time for refresh tokens
func (h *JWTHelper) GetRefreshTokenExpiry() time.Duration {
	return h.refreshTokenExpiry
}
