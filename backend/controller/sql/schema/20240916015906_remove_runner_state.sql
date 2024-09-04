-- migrate:up
ALTER TABLE runners DROP COLUMN "state";
DROP TYPE runner_state;

-- migrate:down

