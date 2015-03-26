CREATE TABLE applications (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    name TEXT NOT NULL
);

CREATE INDEX idx_name on applications (name);

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    kind TEXT NOT NULL,
    email TEXT,
    password TEXT
);

CREATE UNIQUE INDEX idx_email on users (email);

CREATE TABLE applications_users (
    application_id bigint REFERENCES applications (id) ON UPDATE CASCADE ON DELETE CASCADE,
    user_id bigint REFERENCES users (id) ON UPDATE CASCADE ON DELETE CASCADE,
    token TEXT,
    level TEXT,
    CONSTRAINT pidx_application_user PRIMARY KEY (application_id, user_id)
);

CREATE INDEX idx_token on applications_users (token);

CREATE TABLE hosts (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    application_id BIGINT NOT NULL,
    name TEXT NOT NULL,
    tags TEXT[],
    data JSONB
);

CREATE INDEX idx_tags ON hosts USING gin(tags);

CREATE INDEX idx_data ON hosts using gin(data);
