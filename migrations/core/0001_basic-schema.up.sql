CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    email TEXT NOT NULL,
    password TEXT NOT NULL,
    email_verification_token TEXT,
    email_verified BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE UNIQUE INDEX idx_users_email on users (email);
CREATE INDEX idx_users_email_verified on users (email_verified);

CREATE TABLE clusters (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    creator_id BIGSERIAL NOT NULL,
    creator_email TEXT NOT NULL,
    data_retention JSONB NOT NULL DEFAULT '{}',
    members JSONB DEFAULT '[]'
);

CREATE INDEX idx_clusters_name on clusters (name);
CREATE INDEX idx_clusters_creator_id on clusters (creator_id);
CREATE INDEX idx_clusters_creator_email on clusters (creator_email);
CREATE INDEX idx_clusters_members ON clusters USING gin(members);

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
    hostname TEXT NOT NULL,
    updated TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() at time zone 'utc'),
    tags JSONB,
    data JSONB
);

CREATE INDEX idx_hosts_name on hosts (hostname);
CREATE INDEX idx_hosts_tags ON hosts USING gin(tags);
CREATE INDEX idx_hosts_data ON hosts USING gin(data);
CREATE INDEX idx_hosts_updated on hosts (updated);

CREATE TABLE saved_queries (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    user_id bigint REFERENCES users (id),
    cluster_id bigint REFERENCES clusters (id) ON UPDATE CASCADE ON DELETE CASCADE,
    type TEXT NOT NULL,
    query TEXT NOT NULL
);

CREATE INDEX idx_saved_queries_type on saved_queries (type);
CREATE INDEX idx_saved_queries_query on saved_queries (query);

CREATE TABLE metadata (
    cluster_id bigint REFERENCES clusters (id),
    key TEXT NOT NULL,
    data JSONB
);

CREATE INDEX idx_metadata_key on metadata (key);

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
    range TEXT NOT NULL DEFAULT '15 minutes',
    metrics JSONB
);

CREATE INDEX idx_graphs_name on graphs (name);
CREATE INDEX idx_graphs_description on graphs (description);
CREATE INDEX idx_graphs_metrics ON graphs USING gin(metrics);

CREATE TABLE checks (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    cluster_id bigint REFERENCES clusters (id),
    name TEXT NOT NULL,
    interval TEXT NOT NULL,
    is_silenced BOOLEAN NOT NULL DEFAULT FALSE,
    hosts_query TEXT,
    hosts_list JSONB DEFAULT '[]',
    expressions JSONB,
    triggers JSONB,
    last_result_hosts JSONB DEFAULT '[]',
    last_result_expressions JSONB
);

CREATE INDEX idx_checks_name on checks (name);
