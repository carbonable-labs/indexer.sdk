-- name: GetLatestBlock :one
SELECT latest_block FROM indexer_latest_block 
WHERE indexer_hash = $1;

-- name: RecordEvent :exec
INSERT INTO indexer_events (event_id, raw_event, indexer_hash, recorded_at) 
VALUES ($1, $2, $3, $4) 
ON CONFLICT (indexer_hash, event_id) DO UPDATE SET raw_event = $2, recorded_at = $4;
