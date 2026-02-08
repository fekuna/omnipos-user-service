ALTER TABLE merchants ADD COLUMN feature_flags JSONB DEFAULT '{}';
CREATE INDEX idx_merchant_features ON merchants USING GIN (feature_flags);
