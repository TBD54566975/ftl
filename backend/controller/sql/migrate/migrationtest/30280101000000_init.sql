CREATE TABLE test (
  id SERIAL PRIMARY KEY,
  name_and_age TEXT NOT NULL
);

INSERT INTO test (name_and_age) VALUES ('Alice 30');