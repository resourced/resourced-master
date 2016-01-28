CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    email TEXT NOT NULL,
    password TEXT NOT NULL
);

CREATE UNIQUE INDEX idx_users_email on users (email);

CREATE TABLE clusters (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    name TEXT NOT NULL
);

CREATE INDEX idx_clusters_name on clusters (name);

CREATE TABLE clusters_users (
    cluster_id bigint REFERENCES clusters (id) ON UPDATE CASCADE ON DELETE CASCADE,
    user_id bigint REFERENCES users (id) ON UPDATE CASCADE ON DELETE CASCADE,
    -- explicit pk
    CONSTRAINT idx_clusters_users_primary PRIMARY KEY (cluster_id,user_id)
);

CREATE TABLE access_tokens (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    user_id bigint REFERENCES users (id) ON UPDATE CASCADE ON DELETE CASCADE,
    cluster_id bigint REFERENCES clusters (id) ON UPDATE CASCADE ON DELETE CASCADE,
    token TEXT,
    level TEXT,
    enabled BOOLEAN
);

CREATE UNIQUE INDEX idx_access_tokens_token on access_tokens (token);
CREATE INDEX idx_access_tokens_level on access_tokens (level);

CREATE TABLE hosts (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    cluster_id bigint REFERENCES clusters (id) ON UPDATE CASCADE ON DELETE CASCADE,
    access_token_id bigint REFERENCES access_tokens (id),
    name TEXT NOT NULL,
    updated TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() at time zone 'utc'),
    tags JSONB,
    data JSONB
);

CREATE INDEX idx_hosts_name on hosts (name);
CREATE INDEX idx_hosts_tags ON hosts USING gin(tags);
CREATE INDEX idx_hosts_data ON hosts USING gin(data);
CREATE INDEX idx_hosts_updated on hosts (updated);

CREATE TABLE saved_queries (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    user_id bigint REFERENCES users (id),
    cluster_id bigint REFERENCES clusters (id) ON UPDATE CASCADE ON DELETE CASCADE,
    query TEXT NOT NULL
);

CREATE INDEX idx_saved_queries_query on saved_queries (query);

CREATE TABLE metadata (
    cluster_id bigint REFERENCES clusters (id),
    key TEXT NOT NULL,
    data JSONB
);

CREATE INDEX idx_metadata_key on metadata (key);

CREATE TABLE daemons (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    hostname TEXT NOT NULL,
    updated TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() at time zone 'utc')
);

CREATE INDEX idx_daemons_hostname on daemons (hostname);
CREATE INDEX idx_daemons_updated on daemons (updated);

CREATE TABLE watchers (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    cluster_id bigint REFERENCES clusters (id),
    saved_query TEXT,
    name TEXT,
    low_affected_hosts bigint NOT NULL DEFAULT 0,
    hosts_last_updated TEXT,
    check_interval TEXT,
    is_silenced BOOLEAN NOT NULL DEFAULT FALSE,
    active_check JSONB DEFAULT '{}'
);

CREATE INDEX idx_watchers_name on watchers (name);

CREATE TABLE metrics (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    cluster_id bigint REFERENCES clusters (id),
    key TEXT NOT NULL
);

CREATE INDEX idx_metrics_cluster_id_key on metrics (cluster_id, key);

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
