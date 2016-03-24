CREATE TABLE ts_events (
    id bigint NOT NULL,
    cluster_id bigint NOT NULL,
    created_from TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (NOW() at time zone 'utc'),
    created_to TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (NOW() at time zone 'utc'),
    description TEXT NOT NULL
);