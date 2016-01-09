CREATE TABLE ts_watchers (
    cluster_id bigint REFERENCES clusters (id),
    watcher_id bigint REFERENCES watchers (id),
    created TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
    affected_hosts bigint NOT NULL DEFAULT 0,
    data json
);
