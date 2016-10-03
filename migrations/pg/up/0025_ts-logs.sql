CREATE EXTENSION IF NOT EXISTS btree_gin;

CREATE TABLE IF NOT EXISTS ts_logs (
    cluster_id bigint,
    created TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() at time zone 'utc'),
    deleted TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT ((NOW() + interval '30 days') at time zone 'utc'),
    hostname TEXT NOT NULL,
    tags JSONB,
    filename TEXT NOT NULL,
    logline TEXT NOT NULL
);
