CREATE TABLE metadata (
    cluster_id bigint REFERENCES clusters (id),
    key TEXT NOT NULL,
    data JSONB
);

CREATE INDEX idx_metadata_key on metadata (key);
