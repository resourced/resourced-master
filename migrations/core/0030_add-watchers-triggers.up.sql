DROP INDEX IF EXISTS idx_watchers_actions;

ALTER TABLE watchers DROP COLUMN IF EXISTS low_threshold;
ALTER TABLE watchers DROP COLUMN IF EXISTS high_threshold;
ALTER TABLE watchers DROP COLUMN IF EXISTS actions;

CREATE TABLE watchers_triggers (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    cluster_id bigint REFERENCES clusters (id),
    watcher_id bigint REFERENCES watchers (id),
    low_violations_count bigint NOT NULL DEFAULT 0,
    high_violations_count bigint NOT NULL DEFAULT 0,
    created_interval TEXT NOT NULL DEFAULT '24 hour',
    actions JSONB
);

CREATE INDEX idx_watchers_triggers_cluster_id_watcher_id on watchers_triggers (cluster_id, watcher_id);
CREATE INDEX idx_watchers_triggers_actions ON watchers_triggers USING gin(actions);
