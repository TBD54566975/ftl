ALTER TABLE test
  DROP COLUMN name_and_age,
  ALTER COLUMN name SET NOT NULL,
  ALTER COLUMN age SET NOT NULL