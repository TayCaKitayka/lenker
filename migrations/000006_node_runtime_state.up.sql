ALTER TABLE nodes
    ADD COLUMN runtime_mode TEXT NOT NULL DEFAULT 'no-process',
    ADD COLUMN runtime_desired_state TEXT NOT NULL DEFAULT 'validated-config-ready',
    ADD COLUMN runtime_state TEXT NOT NULL DEFAULT 'not_prepared',
    ADD COLUMN last_dry_run_status TEXT NOT NULL DEFAULT 'not_configured',
    ADD COLUMN last_runtime_attempt_status TEXT NOT NULL DEFAULT 'skipped',
    ADD COLUMN last_runtime_prepared_revision INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN last_runtime_transition_at TIMESTAMPTZ,
    ADD COLUMN last_runtime_error TEXT NOT NULL DEFAULT '';

ALTER TABLE nodes
    ADD CONSTRAINT nodes_runtime_mode_check CHECK (runtime_mode IN ('no-process', 'dry-run-only', 'future-process-managed')),
    ADD CONSTRAINT nodes_runtime_state_check CHECK (runtime_state IN ('not_prepared', 'active_config_ready', 'validation_failed', 'prepare_failed')),
    ADD CONSTRAINT nodes_last_dry_run_status_check CHECK (last_dry_run_status IN ('not_configured', 'passed', 'failed')),
    ADD CONSTRAINT nodes_last_runtime_attempt_status_check CHECK (last_runtime_attempt_status IN ('skipped', 'ready', 'failed')),
    ADD CONSTRAINT nodes_last_runtime_prepared_revision_check CHECK (last_runtime_prepared_revision >= 0);
