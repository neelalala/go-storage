DROP TABLE IF EXISTS gc_queue;

DROP TABLE IF EXISTS uploads;

DROP TRIGGER IF EXISTS update_object_metadata_modtime ON objects;

DROP FUNCTION IF EXISTS update_modified_column();

DROP TABLE IF EXISTS objects;
