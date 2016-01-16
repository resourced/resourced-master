CREATE TABLE tasks (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    user_id bigint REFERENCES users (id),
    query TEXT NOT NULL,
    cron TEXT
);

CREATE INDEX idx_tasks_cron on tasks (cron);
CREATE INDEX idx_tasks_query on tasks (query);
