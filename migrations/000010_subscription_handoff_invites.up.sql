CREATE TABLE subscription_handoff_invites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subscription_id UUID NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL DEFAULT 'active',
    expires_at TIMESTAMPTZ NOT NULL,
    claimed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT subscription_handoff_invites_status_check CHECK (status IN ('active', 'claimed', 'revoked')),
    CONSTRAINT subscription_handoff_invites_expires_at_check CHECK (expires_at > created_at)
);

CREATE INDEX subscription_handoff_invites_subscription_id_idx ON subscription_handoff_invites(subscription_id);
CREATE INDEX subscription_handoff_invites_status_expires_at_idx ON subscription_handoff_invites(status, expires_at);
