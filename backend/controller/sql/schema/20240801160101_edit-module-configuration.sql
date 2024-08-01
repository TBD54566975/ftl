-- migrate:up

CREATE UNIQUE INDEX module_name_unique
    ON module_configuration ((COALESCE(module, '')), name);

-- migrate:down

