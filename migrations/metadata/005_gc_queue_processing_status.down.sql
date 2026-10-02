DROP INDEX IF EXISTS idx_gc_queue_status_created_attempts;
UPDATE gc_queue SET status = 'PENDING' WHERE status = 'PROCESSING';
ALTER TABLE gc_queue DROP CONSTRAINT valid_status;
ALTER TABLE gc_queue ADD CONSTRAINT valid_status CHECK (status IN ('PENDING', 'ERROR'));
ALTER TABLE gc_queue DROP COLUMN IF EXISTS updated_at;
