CREATE TABLE IF NOT EXISTS users (
    id bigint primary key,
    email text,
    password text,
    email_verification_token text,
    email_verified boolean
);

CREATE INDEX IF NOT EXISTS idx_users_email on users (email);
CREATE INDEX IF NOT EXISTS idx_users_email_verified on users (email_verified);

CREATE TABLE IF NOT EXISTS clusters (
    id bigint primary key,
    name text,
    creator_id bigint,
    creator_email text,
    data_retention map<text, int>,
    members list<text>
);

CREATE INDEX IF NOT EXISTS idx_clusters_name on clusters (name);
CREATE INDEX IF NOT EXISTS idx_clusters_creator_id on clusters (creator_id);
CREATE INDEX IF NOT EXISTS idx_clusters_creator_email on clusters (creator_email);

CREATE TABLE IF NOT EXISTS access_tokens (
    id bigint primary key,
    user_id bigint,
    cluster_id bigint,
    token_ text,
    level text,
    enabled boolean
);

CREATE INDEX IF NOT EXISTS idx_access_tokens_token on access_tokens (token_);
CREATE INDEX IF NOT EXISTS idx_access_tokens_level on access_tokens (level);

CREATE TABLE IF NOT EXISTS saved_queries (
    id bigint primary key,
    user_id bigint,
    cluster_id bigint,
    type text,
    query text
);

CREATE INDEX IF NOT EXISTS idx_saved_queries_type on saved_queries (type);
CREATE INDEX IF NOT EXISTS idx_saved_queries_query on saved_queries (query);

CREATE TABLE IF NOT EXISTS metrics (
    id bigint primary key,
    cluster_id bigint,
    key text
);

CREATE INDEX IF NOT EXISTS idx_metrics_cluster_id on metrics (cluster_id);
CREATE INDEX IF NOT EXISTS idx_metrics_key on metrics (key);

CREATE TABLE IF NOT EXISTS graphs (
    id bigint primary key,
    cluster_id bigint,
    name text,
    description text,
    range text,
    metrics text
);

CREATE INDEX IF NOT EXISTS idx_graphs_name on graphs (name);
CREATE INDEX IF NOT EXISTS idx_graphs_description on graphs (description);
CREATE INDEX IF NOT EXISTS idx_graphs_metrics ON graphs (metrics);

CREATE TABLE IF NOT EXISTS checks (
    id bigint primary key,
    cluster_id bigint,
    name text,
    interval text,
    is_silenced boolean,
    hosts_query text,
    hosts_list text,
    expressions text,
    triggers text,
    last_result_hosts text,
    last_result_expressions text
);

CREATE INDEX IF NOT EXISTS idx_checks_name on checks (name);
