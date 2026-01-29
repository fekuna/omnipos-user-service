CREATE TABLE refresh_tokens (
	id UUID PRIMARY KEY,
	merchant_id UUID NOT NULL,
	token VARCHAR(512) NOT NULL UNIQUE,
	expires_at TIMESTAMPTZ NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	
	CONSTRAINT fk_merchant
		FOREIGN KEY (merchant_id)
		REFERENCES merchants(id)
		ON DELETE CASCADE
);

-- Create index on token for fast lookups
CREATE INDEX idx_refresh_tokens_token ON refresh_tokens(token);

-- Create index on merchant_id for user token management
CREATE INDEX idx_refresh_tokens_merchant_id ON refresh_tokens(merchant_id);

-- Create index on expires_at for cleanup queries
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
