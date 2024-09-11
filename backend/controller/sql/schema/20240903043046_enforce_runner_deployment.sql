-- migrate:up
DELETE FROM runners WHERE deployment_id IS NULL;
ALTER TABLE runners ALTER COLUMN deployment_id SET NOT NULL;
ALTER TYPE runner_state RENAME VALUE 'idle' TO 'new';
DROP TRIGGER  runners_set_reservation_timeout ON runners;
DROP FUNCTION runners_set_reservation_timeout;
ALTER TABLE runners DROP COLUMN "reservation_timeout";
-- migrate:down

