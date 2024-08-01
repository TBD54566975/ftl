-- migrate:up

ALTER TABLE module_configuration
    ALTER COLUMN module SET DEFAULT '';

-- migrate:down

