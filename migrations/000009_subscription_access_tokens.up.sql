CREATE TABLE subscription_access_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subscription_id UUID NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL DEFAULT 'active',
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT subscription_access_tokens_status_check CHECK (status IN ('active', 'revoked')),
    CONSTRAINT subscription_access_tokens_expires_at_check CHECK (expires_at > created_at)
);

CREATE INDEX subscription_access_tokens_subscription_id_idx ON subscription_access_tokens(subscription_id);
CREATE INDEX subscription_access_tokens_status_expires_at_idx ON subscription_access_tokens(status, expires_at);
