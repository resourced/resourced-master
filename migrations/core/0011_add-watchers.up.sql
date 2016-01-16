CREATE TABLE watchers (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    cluster_id bigint REFERENCES clusters (id),
    saved_query_id bigint REFERENCES saved_queries (id),
    saved_query TEXT NOT NULL,
    name TEXT,
    low_threshold bigint NOT NULL DEFAULT 1,
    high_threshold bigint NOT NULL DEFAULT 0,
    low_affected_hosts bigint NOT NULL DEFAULT 0,
    hosts_last_updated TEXT,
    check_interval TEXT,
    actions JSONB
);

CREATE INDEX idx_watchers_name on watchers (name);
CREATE INDEX idx_watchers_actions ON watchers USING gin(actions);
