-- migrate:up

CREATE UNIQUE INDEX module_config_name_unique
    ON module_configuration ((COALESCE(module, '')), name);

CREATE UNIQUE INDEX module_secret_name_unique
    ON module_secrets ((COALESCE(module, '')), name);

-- migrate:down

