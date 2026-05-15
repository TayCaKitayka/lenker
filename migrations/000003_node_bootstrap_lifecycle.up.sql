ALTER TABLE nodes DROP CONSTRAINT IF EXISTS nodes_status_check;

UPDATE nodes
SET status = CASE status
    WHEN 'registered' THEN 'active'
    WHEN 'healthy' THEN 'active'
    WHEN 'degraded' THEN 'unhealthy'
    WHEN 'offline' THEN 'unhealthy'
    ELSE status
END;

ALTER TABLE nodes
    ALTER COLUMN status SET DEFAULT 'pending',
    ALTER COLUMN auth_token_hash DROP NOT NULL,
    ADD COLUMN registered_at TIMESTAMPTZ,
    ADD COLUMN last_seen_at TIMESTAMPTZ;

UPDATE nodes
SET registered_at = created_at,
    last_seen_at = COALESCE(last_health_at, created_at)
WHERE registered_at IS NULL
  AND auth_token_hash IS NOT NULL;

ALTER TABLE nodes
    ADD CONSTRAINT nodes_status_check CHECK (status IN ('pending', 'active', 'unhealthy', 'drained', 'disabled'));

CREATE TABLE node_bootstrap_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    node_id UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL DEFAULT 'active',
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_by_admin_id UUID REFERENCES admins(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT node_bootstrap_tokens_status_check CHECK (status IN ('active', 'used', 'revoked')),
    CONSTRAINT node_bootstrap_tokens_used_status_check CHECK (
        (status = 'used' AND used_at IS NOT NULL)
        OR (status <> 'used' AND used_at IS NULL)
    )
);

CREATE INDEX node_bootstrap_tokens_node_id_idx ON node_bootstrap_tokens(node_id);
CREATE INDEX node_bootstrap_tokens_status_expires_at_idx ON node_bootstrap_tokens(status, expires_at);
CREATE INDEX nodes_lifecycle_status_idx ON nodes(status, drain_state);
CREATE INDEX nodes_last_seen_at_idx ON nodes(last_seen_at);
