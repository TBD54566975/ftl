-- migrate:up
CREATE TABLE messages( message TEXT );
-- migrate:down
DROP TABLE messages;