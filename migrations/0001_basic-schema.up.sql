CREATE TABLE applications (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    name TEXT NOT NULL
);

CREATE INDEX idx_name on applications (name);

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    application_id BIGINT,
    kind TEXT NOT NULL,
    email TEXT,
    password TEXT,
    token TEXT
);

CREATE INDEX idx_application_id on users (application_id);
CREATE INDEX idx_email on users (email);
CREATE INDEX idx_token on users (token);

CREATE TABLE hosts (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    application_id BIGINT NOT NULL,
    name TEXT NOT NULL,
    tags TEXT[],
    data JSONB
);

CREATE INDEX idx_tags ON hosts USING gin(tags);

CREATE INDEX idx_data ON hosts using gin(data);
