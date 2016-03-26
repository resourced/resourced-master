CREATE EXTENSION IF NOT EXISTS btree_gin;

CREATE TABLE ts_executor_logs (
    cluster_id bigint,
    created TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() at time zone 'utc'),
    hostname TEXT NOT NULL,
    tags JSONB,
    logline TEXT NOT NULL
);

CREATE TABLE ts_logs (
    cluster_id bigint,
    created TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() at time zone 'utc'),
    hostname TEXT NOT NULL,
    tags JSONB,
    filename TEXT NOT NULL,
    logline TEXT NOT NULL
);
