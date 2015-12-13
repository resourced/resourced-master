DROP TABLE IF EXISTS executors CASCADE;
DROP INDEX IF EXISTS idx_executors_hostname;
DROP INDEX IF EXISTS idx_executors_data;

CREATE TABLE tasks (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    user_id bigint REFERENCES users (id),
    query TEXT NOT NULL,
    cron TEXT
);

CREATE INDEX idx_tasks_cron on tasks (cron);
CREATE INDEX idx_tasks_query on tasks (query);
