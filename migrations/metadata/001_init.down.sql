DROP TABLE IF EXISTS uploads;

DROP TABLE IF EXISTS gc_queue;

DROP TRIGGER IF EXISTS update_object_modtime ON objects;

DROP FUNCTION IF EXISTS update_modified_column();

DROP TABLE IF EXISTS objects;
