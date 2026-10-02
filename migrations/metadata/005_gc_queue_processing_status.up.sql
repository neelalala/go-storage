ALTER TABLE gc_queue DROP CONSTRAINT valid_status;
ALTER TABLE gc_queue ADD CONSTRAINT valid_status CHECK (status IN ('PENDING', 'PROCESSING', 'ERROR'));
ALTER TABLE gc_queue ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;
CREATE INDEX IF NOT EXISTS idx_gc_queue_status_created_attempts ON gc_queue (status, created_at, attempts);
