package model

import "time"

type RefreshToken struct {
	BaseModel
	MerchantID string    `db:"merchant_id"`
	Token      string    `db:"token"`
	IsRevoked  bool      `db:"is_revoked"`
	ExpiresAt  time.Time `db:"expires_at"`
}
