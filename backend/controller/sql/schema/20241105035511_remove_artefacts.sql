-- migrate:up
ALTER TABLE deployment_artefacts ADD COLUMN digest BYTEA;
UPDATE deployment_artefacts da SET digest = a.digest FROM artefacts a WHERE da.artefact_id = a.id;
ALTER TABLE deployment_artefacts ALTER COLUMN digest SET NOT NULL;
ALTER TABLE deployment_artefacts DROP COLUMN artefact_id;
ALTER TABLE deployment_artefacts ADD UNIQUE (deployment_id, digest);
DROP TABLE artefacts;
-- migrate:down

