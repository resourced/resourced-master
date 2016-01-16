ALTER TABLE hosts ADD updated TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();

CREATE INDEX idx_updated on hosts (updated);
