CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  display_name TEXT UNIQUE NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  CONSTRAINT valid_name CHECK (
    display_name <> ''
  )
);

INSERT INTO users (id, display_name)
VALUES ('9edd1418-a547-4dd3-971d-709c8cac4b6d', 'system')
ON CONFLICT (id) DO NOTHING;

ALTER TABLE buckets
ADD COLUMN IF NOT EXISTS owner_id UUID REFERENCES users(id) ON DELETE RESTRICT;

UPDATE buckets SET owner_id = '9edd1418-a547-4dd3-971d-709c8cac4b6d' WHERE owner_id IS NULL;

ALTER TABLE buckets
ALTER COLUMN owner_id SET NOT NULL;

ALTER TABLE objects
ADD COLUMN IF NOT EXISTS content_type TEXT NOT NULL DEFAULT 'application/text-plain',
ADD COLUMN IF NOT EXISTS hash TEXT NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS system_metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
ADD COLUMN IF NOT EXISTS user_metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
ADD COLUMN IF NOT EXISTS owner_id UUID REFERENCES users(id) ON DELETE RESTRICT,
DROP COLUMN IF EXISTS checksum;

UPDATE objects SET owner_id = '9edd1418-a547-4dd3-971d-709c8cac4b6d' WHERE owner_id IS NULL;

ALTER TABLE objects
ALTER COLUMN owner_id SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_objects_prefix ON objects (bucket, key text_pattern_ops);

ALTER TABLE uploads
ADD COLUMN IF NOT EXISTS content_type TEXT NOT NULL DEFAULT 'application/octet-stream',
ADD COLUMN IF NOT EXISTS system_metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
ADD COLUMN IF NOT EXISTS user_metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
ADD COLUMN IF NOT EXISTS owner_id UUID REFERENCES users(id) ON DELETE RESTRICT;

UPDATE uploads SET owner_id = '9edd1418-a547-4dd3-971d-709c8cac4b6d' WHERE owner_id IS NULL;
ALTER TABLE uploads ALTER COLUMN owner_id SET NOT NULL;
