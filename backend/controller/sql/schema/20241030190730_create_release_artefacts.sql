-- migrate:up

-- release_artefacts stores references to OCI artefacts (compiled modules)
CREATE TABLE release_artefacts (
    release_id BIGINT      NOT NULL REFERENCES deployments (id) ON DELETE CASCADE,
    digest     BYTEA       UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc'),
    executable BOOLEAN     NOT NULL,
    -- Path relative to the module root.
    path       TEXT        NOT NULL,

    UNIQUE (release_id, digest)
);
-- migrate:down