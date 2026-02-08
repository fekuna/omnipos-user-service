DROP INDEX IF EXISTS idx_merchant_features;
ALTER TABLE merchants DROP COLUMN IF EXISTS feature_flags;
