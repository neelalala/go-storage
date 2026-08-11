CREATE OR REPLACE FUNCTION gc_enqueue_old_object_path()
RETURNS TRIGGER AS $$
BEGIN 
  IF OLD.object_path <> NEW.object_path THEN 
    INSERT INTO gc_queue (object_path, storage_node_id)
    VALUES (OLD.object_path, OLD.storage_node_id);
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';

CREATE TRIGGER enqueue_old_object_path 
  AFTER UPDATE ON objects 
  FOR EACH ROW
  EXECUTE FUNCTION gc_enqueue_old_object_path();
