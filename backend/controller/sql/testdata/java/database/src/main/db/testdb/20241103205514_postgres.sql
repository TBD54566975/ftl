-- migrate:up
CREATE TABLE requests
(
  id SERIAL PRIMARY KEY NOT NULL,
  data TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE SEQUENCE requests_seq;
-- migrate:down
DROP TABLE requests;