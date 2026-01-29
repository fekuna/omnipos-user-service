package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/fekuna/omnipos-user-service/internal/model"
	"github.com/jmoiron/sqlx"
)

type PGRepository struct {
	DB *sqlx.DB
}

func NewPGRepository(db *sqlx.DB) *PGRepository {
	return &PGRepository{DB: db}
}

// Create inserts a new refresh token into the database
func (r *PGRepository) Create(token *model.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, merchant_id, token, is_revoked, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.DB.Exec(
		query,
		token.ID,
		token.MerchantID,
		token.Token,
		token.IsRevoked,
		token.ExpiresAt,
		token.CreatedAt,
	)

	return err
}

// FindByToken retrieves a refresh token by its token string
// Returns nil if token is not found, expired, or revoked
func (r *PGRepository) FindByToken(tokenString string) (*model.RefreshToken, error) {
	var token model.RefreshToken

	query := `
		SELECT id, merchant_id, token, is_revoked, expires_at, created_at, updated_at
		FROM refresh_tokens
		WHERE token = $1 AND expires_at > NOW() AND is_revoked = FALSE
		LIMIT 1
	`

	err := r.DB.Get(&token, query, tokenString)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Token not found, expired, or revoked
		}
		return nil, err
	}

	return &token, nil
}

// RevokeToken marks a specific token as revoked (soft delete)
func (r *PGRepository) RevokeToken(token string) error {
	query := `UPDATE refresh_tokens SET is_revoked = TRUE WHERE token = $1`
	_, err := r.DB.Exec(query, token)
	return err
}

// RevokeAllByMerchantID marks all tokens for a merchant as revoked (soft delete)
// This is useful for logout-all-devices functionality
func (r *PGRepository) RevokeAllByMerchantID(merchantID string) error {
	query := `UPDATE refresh_tokens SET is_revoked = TRUE WHERE merchant_id = $1`
	_, err := r.DB.Exec(query, merchantID)
	return err
}

// DeleteByMerchantID permanently removes all refresh tokens for a specific merchant
// Use RevokeAllByMerchantID for soft deletion instead
func (r *PGRepository) DeleteByMerchantID(merchantID string) error {
	query := `DELETE FROM refresh_tokens WHERE merchant_id = $1`
	_, err := r.DB.Exec(query, merchantID)
	return err
}

// DeleteExpiredTokens removes all expired tokens from the database
// This should be run periodically as a cleanup job
func (r *PGRepository) DeleteExpiredTokens() error {
	query := `DELETE FROM refresh_tokens WHERE expires_at <= $1`
	_, err := r.DB.Exec(query, time.Now())
	return err
}
