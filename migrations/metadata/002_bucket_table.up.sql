CREATE TABLE IF NOT EXISTS bucekts (
  name TEXT PRIMARY KEY,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  CONSTRAINT valid_name CHECK (
    name <> ''
  )
);

INSERT INTO buckets (name, created_at)
SELECT bucket, MIN(created_at)
FROM objects
GROUP BY bucket;

ALTER TABLE objects
ADD CONSTRAINT fk_objects_bucket
FOREIGN KEY (bucket) REFERENCES bucekts(name)
ON DELETE RESTRICT;
