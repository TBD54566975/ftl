-- migrate:up

DROP TABLE identity_keys;
DROP DOMAIN encrypted_identity;

-- migrate:down
