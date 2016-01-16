CREATE TABLE ts_watchers (
    cluster_id bigint,
    watcher_id bigint,
    created TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
    affected_hosts bigint NOT NULL DEFAULT 0,
    data json
);
