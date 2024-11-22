-- migrate:up
CREATE TABLE requests
(
  id SERIAL PRIMARY KEY NOT NULL,
  data TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- migrate:down
DROP TABLE requests;