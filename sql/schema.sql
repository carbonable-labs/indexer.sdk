CREATE TABLE IF NOT EXISTS indexer_events (
  event_id VARCHAR(255) PRIMARY KEY,
  raw_event JSONB NOT NULL DEFAULT '{}',
  recorded_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
  indexer_hash VARCHAR(255) NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS indexer_events_unique_event_id ON indexer_events (indexer_hash, event_id);

CREATE TABLE IF NOT EXISTS indexer_latest_block (
  id INT PRIMARY KEY,
  indexer_hash VARCHAR(255) NOT NULL,
  latest_block VARCHAR(255) NOT NULL
);
