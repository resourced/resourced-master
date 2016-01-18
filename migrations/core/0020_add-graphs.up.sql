CREATE TABLE graphs (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    cluster_id bigint REFERENCES clusters (id),
    name TEXT NOT NULL,
    description TEXT,
    metrics JSONB
);

CREATE INDEX idx_graphs_name on graphs (name);
CREATE INDEX idx_graphs_description on graphs (description);
CREATE INDEX idx_graphs_metrics ON graphs USING gin(metrics);
