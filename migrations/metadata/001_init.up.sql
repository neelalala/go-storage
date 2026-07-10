CREATE TABLE IF NOT EXISTS objects (
  bucket TEXT NOT NULL,
  key TEXT NOT NULL,
  object_path TEXT NOT NULL,
  size BIGINT NOT NULL,
  checksum BIGINT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  storage_node_id UUID NOT NULL, 

  PRIMARY KEY (bucket, key),

  CONSTRAINT valid_bucket CHECK (
    bucket <> ''  
  ),

  CONSTRAINT valid_key CHECK (
    key <> ''
  ),

  CONSTRAINT valid_object_path CHECK (
    object_path <> ''
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

CREATE TABLE IF NOT EXISTS uploads (
  upload_id UUID PRIMARY KEY,
  bucket TEXT NOT NULL,
  key TEXT NOT NULL,
  object_path TEXT NOT NULL,
  size BIGINT NOT NULL,
  storage_node_id UUID NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  CONSTRAINT valid_bucket CHECK (
    bucket <> ''  
  ),

  CONSTRAINT valid_key CHECK (
    key <> ''
  ),

  CONSTRAINT valid_object_path CHECK (
    object_path <> ''
  )
);


CREATE TABLE IF NOT EXISTS gc_queue (
  deletion_id BIGSERIAL PRIMARY KEY,
  object_path TEXT NOT NULL,
  storage_node_id UUID NOT NULL,
  status TEXT NOT NULL DEFAULT 'PENDING',
  attempts INT NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  CONSTRAINT valid_object_path CHECK (
    object_path <> ''
  ),

  CONSTRAINT valid_status CHECK (
    status IN ('PENDING', 'ERROR')
  )
);
