CREATE TABLE IF NOT EXISTS object_metadata (
  bucket TEXT NOT NULL,
  key TEXT NOT NULL,
  size BIGINT NOT NULL,
  chechsum BIGINT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  PRIMARY KEY (bucket, key),

  CONSTRAINT valid_bucket CHECK (
    bucket <> ''  
  ),

  CONSTRAINT valid_key CHECK (
    key <> ''
  )
);

CREATE OR REPLACE FUNCTION update_modified_column()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_object_metadata_modtime
  BEFORE UPDATE ON object_metadata
  FOR EACH ROW
  EXECUTE PROCEDURE update_modified_column();
