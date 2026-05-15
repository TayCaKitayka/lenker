DROP INDEX IF EXISTS nodes_last_seen_at_idx;
DROP INDEX IF EXISTS nodes_lifecycle_status_idx;
DROP INDEX IF EXISTS node_bootstrap_tokens_status_expires_at_idx;
DROP INDEX IF EXISTS node_bootstrap_tokens_node_id_idx;

DROP TABLE IF EXISTS node_bootstrap_tokens;

ALTER TABLE nodes DROP CONSTRAINT IF EXISTS nodes_status_check;

UPDATE nodes
SET status = CASE status
    WHEN 'pending' THEN 'registered'
    WHEN 'active' THEN 'healthy'
    WHEN 'unhealthy' THEN 'degraded'
    WHEN 'drained' THEN 'healthy'
    WHEN 'disabled' THEN 'offline'
    ELSE status
END,
    last_health_at = COALESCE(last_health_at, last_seen_at);

UPDATE nodes
SET auth_token_hash = 'down-migration-placeholder-' || id::text
WHERE auth_token_hash IS NULL;

ALTER TABLE nodes
    ALTER COLUMN status SET DEFAULT 'registered',
    ALTER COLUMN auth_token_hash SET NOT NULL,
    DROP COLUMN IF EXISTS last_seen_at,
    DROP COLUMN IF EXISTS registered_at;

ALTER TABLE nodes
    ADD CONSTRAINT nodes_status_check CHECK (status IN ('registered', 'healthy', 'degraded', 'offline'));
