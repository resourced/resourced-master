DROP TABLE IF EXISTS clusters CASCADE;
DROP INDEX IF EXISTS idx_clusters_name;

DROP TABLE IF EXISTS clusters_users CASCADE;

ALTER TABLE access_tokens DROP cluster_id;
