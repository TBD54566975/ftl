-- name: UpsertTopic :exec
INSERT INTO topics (key, module_name, name, type)
VALUES (
           sqlc.arg('topic')::topic_key,
           sqlc.arg('module')::TEXT,
           sqlc.arg('name')::TEXT,
           sqlc.arg('event_type')::TEXT
       )
ON CONFLICT (name, module_name) DO
    UPDATE SET
    type = sqlc.arg('event_type')::TEXT
RETURNING id;

-- name: UpsertSubscription :one
INSERT INTO topic_subscriptions (
    key,
    topic_id,
    module_name,
    deployment_key,
    name)
VALUES (
           sqlc.arg('key')::subscription_key,
           (
               SELECT topics.id as id
               FROM topics
               WHERE module_name = sqlc.arg('topic_module')::TEXT
                 AND topics.name = sqlc.arg('topic_name')::TEXT
           ),
           sqlc.arg('module')::TEXT,
           sqlc.arg('deployment')::deployment_key,
           sqlc.arg('name')::TEXT
       )
ON CONFLICT (name, module_name) DO
    UPDATE SET
               topic_id = excluded.topic_id,
               deployment_key = sqlc.arg('deployment')::deployment_key
RETURNING
    id,
    CASE
        WHEN xmax = 0 THEN true
        ELSE false
        END AS inserted;

-- name: DeleteSubscriptions :many
DELETE FROM topic_subscriptions
WHERE deployment_key = sqlc.arg('deployment')::deployment_key
RETURNING topic_subscriptions.key;

-- name: DeleteSubscribers :many
DELETE FROM topic_subscribers
WHERE deployment_key = sqlc.arg('deployment')::deployment_key
RETURNING topic_subscribers.key;

-- name: InsertSubscriber :exec
INSERT INTO topic_subscribers (
    key,
    topic_subscriptions_id,
    deployment_key,
    sink,
    retry_attempts,
    backoff,
    max_backoff,
    catch_verb
)
VALUES (
           sqlc.arg('key')::subscriber_key,
           (
               SELECT topic_subscriptions.id as id
               FROM topic_subscriptions
               WHERE module_name = sqlc.arg('module')::TEXT
                 AND topic_subscriptions.name = sqlc.arg('subscription_name')::TEXT
           ),
           sqlc.arg('deployment')::deployment_key,
           sqlc.arg('sink'),
           sqlc.arg('retry_attempts'),
           sqlc.arg('backoff')::interval,
           sqlc.arg('max_backoff')::interval,
           sqlc.arg('catch_verb')
       );

-- name: PublishEventForTopic :exec
INSERT INTO topic_events (
    "key",
    topic_id,
    caller,
    payload,
    request_key,
    trace_context
)
VALUES (
           sqlc.arg('key')::topic_event_key,
           (
               SELECT topics.id
               FROM topics
               WHERE module_name = sqlc.arg('module')::TEXT
                 AND topics.name = sqlc.arg('topic')::TEXT
           ),
           sqlc.arg('caller')::TEXT,
           sqlc.arg('payload'),
           sqlc.arg('request_key')::TEXT,
           sqlc.arg('trace_context')::jsonb
       );

-- name: GetSubscriptionsNeedingUpdate :many
-- Results may not be ready to be scheduled yet due to event consumption delay
-- Sorting ensures that brand new events (that may not be ready for consumption)
-- don't prevent older events from being consumed
-- We also make sure that the subscription belongs to a deployment that has at least one runner
SELECT
    subs.key::subscription_key as key,
    curser.key as cursor,
    topics.key::topic_key as topic,
    subs.name,
    deployment_key as deployment_key,
    curser.request_key as request_key
FROM topic_subscriptions subs
         LEFT JOIN topics ON subs.topic_id = topics.id
         LEFT JOIN topic_events curser ON subs.cursor = curser.id
WHERE subs.cursor IS DISTINCT FROM topics.head
  AND subs.state = 'idle'
ORDER BY curser.created_at
LIMIT 3
    FOR UPDATE OF subs SKIP LOCKED;


-- name: GetNextEventForSubscription :one
WITH cursor AS (
    SELECT
        created_at,
        id
    FROM topic_events
    WHERE "key" = sqlc.narg('cursor')::topic_event_key
)
SELECT events."key" as event,
       events.payload,
       events.created_at,
       events.caller,
       events.request_key,
       events.trace_context,
       NOW() - events.created_at >= sqlc.arg('consumption_delay')::interval AS ready
FROM topics
         LEFT JOIN topic_events as events ON events.topic_id = topics.id
WHERE topics.key = sqlc.arg('topic')::topic_key
  AND (events.created_at, events.id) > (SELECT COALESCE(MAX(cursor.created_at), '1900-01-01'), COALESCE(MAX(cursor.id), 0) FROM cursor)
ORDER BY events.created_at, events.id
LIMIT 1;

-- name: GetRandomSubscriber :one
SELECT
    subscribers.sink as sink,
    subscribers.retry_attempts as retry_attempts,
    subscribers.backoff as backoff,
    subscribers.max_backoff as max_backoff,
    subscribers.catch_verb as catch_verb,
    subscribers.deployment_key as deployment_key
FROM topic_subscribers as subscribers
         JOIN topic_subscriptions ON subscribers.topic_subscriptions_id = topic_subscriptions.id
WHERE topic_subscriptions.key = sqlc.arg('key')::subscription_key
ORDER BY RANDOM()
LIMIT 1;

-- name: BeginConsumingTopicEvent :exec
WITH event AS (
    SELECT *
    FROM topic_events
    WHERE "key" = sqlc.arg('event')::topic_event_key
)
UPDATE topic_subscriptions
SET state = 'executing',
    cursor = (SELECT id FROM event)
WHERE key = sqlc.arg('subscription')::subscription_key;

-- name: CompleteEventForSubscription :exec
UPDATE topic_subscriptions
SET state = 'idle'
WHERE name = @name::TEXT
  AND module_name = sqlc.arg('module')::TEXT;

-- name: GetSubscription :one
SELECT *
FROM topic_subscriptions
WHERE name = $1::TEXT
  AND module_name = $2::TEXT;

-- name: SetSubscriptionCursor :exec
WITH event AS (
    SELECT id, created_at, key, topic_id, payload
    FROM topic_events
    WHERE "key" = $2::topic_event_key
)
UPDATE topic_subscriptions
SET cursor = (SELECT id FROM event)
WHERE key = $1::subscription_key;

-- name: GetTopic :one
SELECT *
FROM topics
WHERE id = $1::BIGINT;

-- name: GetTopicEvent :one
SELECT *
FROM topic_events
WHERE id = $1::BIGINT;

