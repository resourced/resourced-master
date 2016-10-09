CREATE TABLE IF NOT EXISTS ts_metrics (
    cluster_id bigint,
    metric_id bigint,
    created TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() at time zone 'utc'),
    deleted TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT ((NOW() + interval '30 days') at time zone 'utc'),
    key TEXT NOT NULL,
    host TEXT NOT NULL,
    value double precision NOT NULL DEFAULT 0
);
