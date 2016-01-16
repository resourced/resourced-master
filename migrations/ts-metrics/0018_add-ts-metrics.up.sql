CREATE TABLE ts_metrics (
    cluster_id bigint,
    metric_id bigint,
    created TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
    key TEXT NOT NULL,
    host TEXT NOT NULL,
    value double precision NOT NULL DEFAULT 0
);
