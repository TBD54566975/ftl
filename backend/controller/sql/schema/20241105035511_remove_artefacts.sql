-- migrate:up
ALTER TABLE deployment_artefacts DROP COLUMN artefact_id;
ALTER TABLE deployment_artefacts ADD COLUMN digest BYTEA NOT NULL;
ALTER TABLE deployment_artefacts ADD UNIQUE (deployment_id, digest);
DROP TABLE artefacts;
-- migrate:down

