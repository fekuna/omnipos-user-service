ALTER TABLE refresh_tokens ADD COLUMN user_id UUID REFERENCES users(id) ON DELETE CASCADE;
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
