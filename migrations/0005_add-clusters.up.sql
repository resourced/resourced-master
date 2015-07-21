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

ALTER TABLE access_tokens ADD cluster_id bigint REFERENCES clusters (id) ON UPDATE CASCADE ON DELETE CASCADE;
