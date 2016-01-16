CREATE TABLE saved_queries (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    user_id bigint REFERENCES users (id),
    query TEXT NOT NULL
);

CREATE INDEX idx_saved_queries_query on saved_queries (query);
