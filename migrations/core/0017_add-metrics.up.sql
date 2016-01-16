CREATE TABLE metrics (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    cluster_id bigint REFERENCES clusters (id),
    key TEXT NOT NULL
);

CREATE INDEX idx_metrics_cluster_id_key on metrics (cluster_id, key);
