-- migrate:up
CREATE TABLE requests (
     id bigint NOT NULL,
     created_at datetime(6),
     data VARCHAR(255),
     updated_at datetime(6),
     PRIMARY KEY (id)
 ) engine=InnoDB;

CREATE TABLE requests_SEQ (
     next_val bigint
 ) engine=InnoDB;

INSERT INTO requests_SEQ VALUES ( 1 );

-- migrate:down
DROP TABLE requests;