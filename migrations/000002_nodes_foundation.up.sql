CREATE TABLE nodes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL DEFAULT '',
    region TEXT NOT NULL DEFAULT '',
    country_code TEXT NOT NULL DEFAULT '',
    hostname TEXT NOT NULL DEFAULT '',
    public_ipv4 TEXT,
    public_ipv6 TEXT,
    status TEXT NOT NULL DEFAULT 'registered',
    drain_state TEXT NOT NULL DEFAULT 'active',
    agent_version TEXT NOT NULL DEFAULT '',
    xray_version TEXT NOT NULL DEFAULT '',
    auth_token_hash TEXT NOT NULL UNIQUE,
    active_revision INTEGER NOT NULL DEFAULT 0,
    last_health_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT nodes_status_check CHECK (status IN ('registered', 'healthy', 'degraded', 'offline')),
    CONSTRAINT nodes_drain_state_check CHECK (drain_state IN ('active', 'draining', 'drained'))
);

CREATE TABLE node_registrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    node_id UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    bootstrap_token_hash TEXT NOT NULL,
    registration_status TEXT NOT NULL DEFAULT 'completed',
    certificate_fingerprint TEXT,
    registered_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT node_registrations_status_check CHECK (registration_status IN ('completed', 'rejected'))
);

CREATE INDEX nodes_status_idx ON nodes(status);
CREATE INDEX nodes_last_health_at_idx ON nodes(last_health_at);
CREATE INDEX node_registrations_node_id_idx ON node_registrations(node_id);
