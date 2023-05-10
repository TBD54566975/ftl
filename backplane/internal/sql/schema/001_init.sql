-- +goose Up
CREATE EXTENSION "uuid-ossp";

CREATE TABLE modules (
  id BIGSERIAL PRIMARY KEY, 
  language VARCHAR(64) NOT NULL,
  name VARCHAR(128) UNIQUE NOT NULL
);

CREATE TABLE deployments (
  id BIGSERIAL PRIMARY KEY,
  created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc'),
  module_id BIGINT NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
  -- Unique identifier for this deployment.
  "key" UUID UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
  -- Proto-encoded module schema.
  "schema" BYTEA NOT NULL
);

CREATE INDEX deployments_module_id_idx ON deployments (module_id);

CREATE TABLE artefacts (
  id BIGSERIAL PRIMARY KEY NOT NULL,
  created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc'),
  -- SHA256 digest of the content.
  digest BYTEA UNIQUE NOT NULL,
  content BYTEA NOT NULL
);

CREATE TABLE deployment_artefacts (
  artefact_id BIGINT NOT NULL REFERENCES artefacts(id) ON DELETE CASCADE,
  deployment_id BIGINT NOT NULL REFERENCES deployments(id) ON DELETE CASCADE,
  created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc'),
  executable BOOLEAN NOT NULL,
  -- Path relative to the module root.
  path VARCHAR(128) NOT NULL
);

CREATE INDEX deployment_artefacts_deployment_artefact_idx ON deployment_artefacts (artefact_id, deployment_id);