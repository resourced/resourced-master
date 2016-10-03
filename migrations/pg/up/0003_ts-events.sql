CREATE TABLE IF NOT EXISTS ts_events (
    id bigint NOT NULL,
    cluster_id bigint NOT NULL,
    created_from TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() at time zone 'utc'),
    created_to TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() at time zone 'utc'),
    deleted TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT ((NOW() + interval '30 days') at time zone 'utc'),
    description TEXT NOT NULL
);