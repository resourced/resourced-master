CREATE TABLE daemons (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    hostname TEXT NOT NULL,
    updated TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_daemons_hostname on daemons (hostname);
CREATE INDEX idx_daemons_updated on daemons (updated);
