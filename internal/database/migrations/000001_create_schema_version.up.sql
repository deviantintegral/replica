-- Initial migration to verify migrations work
CREATE TABLE IF NOT EXISTS replica_metadata (
    key VARCHAR(255) PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO replica_metadata (key, value) VALUES ('schema_version', '1');
