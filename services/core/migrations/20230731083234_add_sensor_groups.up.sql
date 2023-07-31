CREATE TABLE sensor_groups (
    id BIGINT NOT NULL GENERATED ALWAYS AS IDENTITY,
    name varchar NOT NULL,
    description varchar NOT NULL DEFAULT(''),

    PRIMARY KEY(id)
);

CREATE TABLE sensor_groups_sensors (
    sensor_group_id BIGINT NOT NULL REFERENCES sensor_groups(id),
    sensor_id BIGINT NOT NULL REFERENCES sensors(id),

    UNIQUE(sensor_group_id, sensor_id)
);
