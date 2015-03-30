CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    email TEXT NOT NULL,
    password TEXT NOT NULL
);

CREATE UNIQUE INDEX idx_email on users (email);

CREATE TABLE access_tokens (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    user_id bigint REFERENCES users (id) ON UPDATE CASCADE ON DELETE CASCADE,
    token TEXT,
    level TEXT
);

CREATE UNIQUE INDEX idx_token on access_tokens (token);
CREATE INDEX idx_level on access_tokens (level);

CREATE TABLE hosts (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    access_token_id bigint REFERENCES access_tokens (id),
    name TEXT NOT NULL,
    tags TEXT[],
    data JSONB
);

CREATE INDEX idx_tags ON hosts USING gin(tags);

CREATE INDEX idx_data ON hosts USING gin(data);
