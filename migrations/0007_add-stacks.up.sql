CREATE TABLE stacks (
    cluster_id bigint REFERENCES clusters (id),
    data JSONB
);

CREATE INDEX idx_stacks_data ON stacks USING gin(data);
