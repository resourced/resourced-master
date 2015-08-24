CREATE TABLE executors (
    cluster_id bigint REFERENCES clusters (id),
    hostname TEXT NOT NULL,
    data JSONB
);

CREATE INDEX idx_executors_hostname on executors (hostname);
CREATE INDEX idx_executors_data ON executors USING gin(data);
