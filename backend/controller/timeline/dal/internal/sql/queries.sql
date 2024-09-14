-- name: InsertTimelineLogEvent :exec
INSERT INTO timeline (
  deployment_id,
  request_id,
  time_stamp,
  custom_key_1,
  type,
  payload
)
VALUES (
  (SELECT id FROM deployments d WHERE d.key = sqlc.arg('deployment_key')::deployment_key LIMIT 1),
  (
    CASE
      WHEN sqlc.narg('request_key')::TEXT IS NULL THEN NULL
      ELSE (SELECT id FROM requests ir WHERE ir.key = sqlc.narg('request_key')::TEXT LIMIT 1)
    END
  ),
  sqlc.arg('time_stamp')::TIMESTAMPTZ,
  sqlc.arg('level')::INT,
  'log',
  sqlc.arg('payload')
);

-- name: InsertTimelineCallEvent :exec
INSERT INTO timeline (
  deployment_id,
  request_id,
  parent_request_id,
  time_stamp,
  type,
  custom_key_1,
  custom_key_2,
  custom_key_3,
  custom_key_4,
  payload
)
VALUES (
  (SELECT id FROM deployments WHERE deployments.key = sqlc.arg('deployment_key')::deployment_key),
  (CASE
      WHEN sqlc.narg('request_key')::TEXT IS NULL THEN NULL
      ELSE (SELECT id FROM requests ir WHERE ir.key = sqlc.narg('request_key')::TEXT)
    END),
  (CASE
      WHEN sqlc.narg('parent_request_key')::TEXT IS NULL THEN NULL
      ELSE (SELECT id FROM requests ir WHERE ir.key = sqlc.narg('parent_request_key')::TEXT)
    END),
  sqlc.arg('time_stamp')::TIMESTAMPTZ,
  'call',
  sqlc.narg('source_module')::TEXT,
  sqlc.narg('source_verb')::TEXT,
  sqlc.arg('dest_module')::TEXT,
  sqlc.arg('dest_verb')::TEXT,
  sqlc.arg('payload')
);

-- name: DeleteOldTimelineEvents :one
WITH deleted AS (
    DELETE FROM timeline
    WHERE time_stamp < (NOW() AT TIME ZONE 'utc') - sqlc.arg('timeout')::INTERVAL
      AND type = sqlc.arg('type')
    RETURNING 1
)
SELECT COUNT(*)
FROM deleted;

-- name: DummyQueryTimeline :one
-- This is a dummy query to ensure that the Timeline model is generated.
SELECT * FROM timeline WHERE id = @id;
