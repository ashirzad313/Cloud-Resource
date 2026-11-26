-- WARNING: This rollback removes the entire virtual_machines table
-- AND ALL OF ITS DATA. Do not run unless you intend to destroy it.
BEGIN;

DROP TABLE virtual_machines;

COMMIT;
