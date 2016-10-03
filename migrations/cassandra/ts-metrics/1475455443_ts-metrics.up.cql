CREATE TABLE ts_metrics (
    cluster_id bigint NOT NULL,
    metric_id bigint NOT NULL,
    created bigint NOT NULL,
    key text NOT NULL,
    host text,
    value double double DEFAULT 0
);

CREATE INDEX ON ts_metrics (cluster_id, metric_id, created, key, host);

CREATE TABLE ts_metrics_aggr_15m (
    cluster_id bigint,
    metric_id bigint,
    created bigint,
    key text NOT NULL,
    host text,
    avg double DEFAULT 0,
    max double DEFAULT 0,
    min double DEFAULT 0,
    sum double DEFAULT 0
);

CREATE INDEX ON ts_metrics_aggr_15m (cluster_id, metric_id, created, key, host);
