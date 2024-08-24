-- migrate:up
DELETE FROM runners WHERE deployment_id IS NULL;
ALTER TABLE runners ALTER COLUMN deployment_id SET NOT NULL;
ALTER TYPE runner_state RENAME VALUE 'idle' TO 'new';
-- migrate:down

