BEGIN;

CREATE TABLE user_workers (
    id UUID PRIMARY KEY NOT NULL,
    name VARCHAR NOT NULL,
    description VARCHAR NOT NULL DEFAULT(''),
    state VARCHAR NOT NULL DEFAULT('disabled'),
    language VARCHAR NOT NULL,
    organisation BIGINT NOT NULL,
    revision INT NOT NULL,
    status VARCHAR NOT NULL,
    status_info TEXT NOT NULL DEFAULT(''),
    source BYTEA NOT NULL,
    entrypoint VARCHAR NOT NULL
);

COMMIT;
