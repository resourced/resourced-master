CREATE TABLE watchers_results (
    cluster_id bigint REFERENCES clusters (id),
    watcher_id bigint REFERENCES watchers (id),
    affected_hosts bigint NOT NULL DEFAULT 0,
    count bigint NOT NULL DEFAULT 0,
    updated TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_watchers_results_affected_hosts on watchers_results (affected_hosts);
CREATE INDEX idx_watchers_results_actions ON watchers_results (count);
CREATE INDEX idx_watchers_results_updated on watchers_results (updated);
