CREATE TABLE IF NOT EXISTS hosts (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    cluster_id bigint NOT NULL,
    access_token_id bigint NOT NULL,
    hostname TEXT NOT NULL,
    updated TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() at time zone 'utc'),
    tags JSONB,
    master_tags JSONB DEFAULT '{}',
    data JSONB
);

CREATE INDEX IF NOT EXISTS idx_hosts_cluster_id on hosts (cluster_id);
CREATE INDEX IF NOT EXISTS idx_hosts_name on hosts (hostname);
CREATE INDEX IF NOT EXISTS idx_hosts_tags ON hosts USING gin(tags);
CREATE INDEX IF NOT EXISTS idx_hosts_master_tags ON hosts USING gin(master_tags);
CREATE INDEX IF NOT EXISTS idx_hosts_data ON hosts USING gin(data);
CREATE INDEX IF NOT EXISTS idx_hosts_updated on hosts (updated);
