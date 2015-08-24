ALTER TABLE saved_queries ADD cluster_id bigint REFERENCES clusters (id) ON UPDATE CASCADE ON DELETE CASCADE;
