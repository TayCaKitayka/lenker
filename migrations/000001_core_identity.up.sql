CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE admins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    two_factor_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_login_at TIMESTAMPTZ,
    CONSTRAINT admins_status_check CHECK (status IN ('active', 'suspended'))
);

CREATE TABLE admin_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT ''
);

CREATE TABLE admin_role_bindings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_id UUID NOT NULL REFERENCES admins(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES admin_roles(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (admin_id, role_id)
);

CREATE TABLE admin_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_id UUID NOT NULL REFERENCES admins(id) ON DELETE CASCADE,
    session_token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL DEFAULT 'active',
    display_name TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT users_status_check CHECK (status IN ('active', 'suspended', 'expired'))
);

CREATE TABLE user_identities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type TEXT NOT NULL,
    email TEXT NOT NULL,
    email_verified_at TIMESTAMPTZ,
    magic_link_nonce_hash TEXT,
    last_auth_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT user_identities_type_check CHECK (type = 'email'),
    UNIQUE (type, email)
);

CREATE TABLE plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    duration_days INTEGER NOT NULL,
    traffic_limit_bytes BIGINT,
    device_limit INTEGER NOT NULL,
    region_policy JSONB NOT NULL DEFAULT '{}'::jsonb,
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT plans_duration_days_check CHECK (duration_days > 0),
    CONSTRAINT plans_device_limit_check CHECK (device_limit > 0),
    CONSTRAINT plans_traffic_limit_bytes_check CHECK (traffic_limit_bytes IS NULL OR traffic_limit_bytes > 0),
    CONSTRAINT plans_status_check CHECK (status IN ('active', 'archived'))
);

CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan_id UUID NOT NULL REFERENCES plans(id),
    status TEXT NOT NULL DEFAULT 'active',
    starts_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    traffic_limit_bytes BIGINT,
    traffic_used_bytes BIGINT NOT NULL DEFAULT 0,
    device_limit INTEGER NOT NULL,
    preferred_region TEXT,
    last_key_rotated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT subscriptions_status_check CHECK (status IN ('active', 'expired', 'suspended')),
    CONSTRAINT subscriptions_period_check CHECK (expires_at > starts_at),
    CONSTRAINT subscriptions_device_limit_check CHECK (device_limit > 0),
    CONSTRAINT subscriptions_traffic_limit_bytes_check CHECK (traffic_limit_bytes IS NULL OR traffic_limit_bytes > 0),
    CONSTRAINT subscriptions_traffic_used_bytes_check CHECK (traffic_used_bytes >= 0)
);

CREATE INDEX admin_sessions_admin_id_idx ON admin_sessions(admin_id);
CREATE INDEX user_identities_user_id_idx ON user_identities(user_id);
CREATE INDEX subscriptions_user_id_idx ON subscriptions(user_id);
CREATE INDEX subscriptions_plan_id_idx ON subscriptions(plan_id);
CREATE INDEX subscriptions_status_idx ON subscriptions(status);
