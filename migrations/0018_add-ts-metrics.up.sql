CREATE TABLE ts_metrics (
    cluster_id bigint REFERENCES clusters (id),
    metric_id bigint REFERENCES metrics (id),
    created TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
    key TEXT NOT NULL,
    host TEXT NOT NULL,
    value bigint NOT NULL DEFAULT 0
);
