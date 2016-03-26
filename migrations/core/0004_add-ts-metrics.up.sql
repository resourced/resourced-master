CREATE TABLE ts_metrics (
    cluster_id bigint,
    metric_id bigint,
    created TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() at time zone 'utc'),
    key TEXT NOT NULL,
    host TEXT NOT NULL,
    value double precision NOT NULL DEFAULT 0
);

CREATE TABLE ts_metrics_aggr_15m (
    cluster_id bigint,
    metric_id bigint,
    created TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() at time zone 'utc'),
    key TEXT NOT NULL,
    host TEXT,
    avg double precision NOT NULL DEFAULT 0,
    max double precision NOT NULL DEFAULT 0,
    min double precision NOT NULL DEFAULT 0,
    sum double precision NOT NULL DEFAULT 0
);
