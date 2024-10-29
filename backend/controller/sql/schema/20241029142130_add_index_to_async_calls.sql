-- migrate:up

CREATE INDEX idx_async_calls_lease_id ON async_calls(lease_id);

-- migrate:down

