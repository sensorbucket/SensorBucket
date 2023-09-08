BEGIN;

CREATE TABLE user_workers (
    id UUID PRIMARY KEY NOT NULL,
    name VARCHAR NOT NULL,
    description VARCHAR NOT NULL DEFAULT(''),
    state INT NOT NULL DEFAULT(0),
    language INT NOT NULL,
    organisation BIGINT NOT NULL,
    major INT NOT NULL,
    revision INT NOT NULL,
    status INT NOT NULL,
    status_info TEXT NOT NULL DEFAULT(''),
    source BYTEA NOT NULL,
    entrypoint VARCHAR NOT NULL
);

COMMIT;
