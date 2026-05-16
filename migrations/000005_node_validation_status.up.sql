ALTER TABLE nodes
    ADD COLUMN last_validation_status TEXT NOT NULL DEFAULT '',
    ADD COLUMN last_validation_error TEXT NOT NULL DEFAULT '',
    ADD COLUMN last_validation_at TIMESTAMPTZ,
    ADD COLUMN last_applied_revision INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN active_config_path TEXT NOT NULL DEFAULT '';

ALTER TABLE nodes
    ADD CONSTRAINT nodes_last_validation_status_check CHECK (
        last_validation_status IN ('', 'applied', 'failed')
    ),
    ADD CONSTRAINT nodes_last_applied_revision_check CHECK (last_applied_revision >= 0);
