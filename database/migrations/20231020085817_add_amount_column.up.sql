BEGIN TRANSACTION;

ALTER TABLE orders ADD amount TEXT;

COMMIT;