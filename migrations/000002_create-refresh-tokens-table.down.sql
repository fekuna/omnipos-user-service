DROP INDEX IF EXISTS idx_refresh_tokens_expires_at;
DROP INDEX IF EXISTS idx_refresh_tokens_merchant_id;
DROP INDEX IF EXISTS idx_refresh_tokens_token;
DROP TABLE IF EXISTS refresh_tokens CASCADE;
