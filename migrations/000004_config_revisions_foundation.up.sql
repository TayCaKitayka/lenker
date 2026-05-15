CREATE TABLE config_revisions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    node_id UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    revision_number INTEGER NOT NULL,
    bundle_hash TEXT NOT NULL,
    signature TEXT NOT NULL,
    signer TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    rollback_target_revision INTEGER,
    bundle_json JSONB NOT NULL,
    created_by_admin_id UUID REFERENCES admins(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    applied_at TIMESTAMPTZ,
    failed_at TIMESTAMPTZ,
    rolled_back_at TIMESTAMPTZ,
    error_message TEXT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT config_revisions_status_check CHECK (status IN ('pending', 'applied', 'failed', 'rolled_back')),
    CONSTRAINT config_revisions_revision_number_check CHECK (revision_number > 0),
    CONSTRAINT config_revisions_rollback_target_check CHECK (
        rollback_target_revision IS NULL OR rollback_target_revision >= 0
    ),
    CONSTRAINT config_revisions_node_revision_unique UNIQUE (node_id, revision_number)
);

CREATE INDEX config_revisions_node_id_status_idx ON config_revisions(node_id, status);
CREATE INDEX config_revisions_node_id_created_at_idx ON config_revisions(node_id, created_at DESC);
