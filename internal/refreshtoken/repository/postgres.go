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
		INSERT INTO refresh_tokens (id, merchant_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.DB.Exec(
		query,
		token.ID,
		token.MerchantID,
		token.Token,
		token.ExpiresAt,
		token.CreatedAt,
	)

	return err
}

// FindByToken retrieves a refresh token by its token string
func (r *PGRepository) FindByToken(tokenString string) (*model.RefreshToken, error) {
	var token model.RefreshToken

	query := `
		SELECT id, merchant_id, token, expires_at, created_at
		FROM refresh_tokens
		WHERE token = $1 AND expires_at > NOW()
		LIMIT 1
	`

	err := r.DB.Get(&token, query, tokenString)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Token not found or expired
		}
		return nil, err
	}

	return &token, nil
}

// DeleteByMerchantID removes all refresh tokens for a specific merchant
// This is useful for logout functionality or when revoking all sessions
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
