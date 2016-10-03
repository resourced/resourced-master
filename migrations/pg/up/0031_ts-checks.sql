CREATE TABLE IF NOT EXISTS ts_checks (
    cluster_id bigint,
    check_id bigint,
    created TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() at time zone 'utc'),
    deleted TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT ((NOW() + interval '30 days') at time zone 'utc'),
    result BOOLEAN NOT NULL DEFAULT FALSE,
    expressions jsonb
);
