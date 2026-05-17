ALTER TABLE nodes
    ADD COLUMN runtime_process_mode TEXT NOT NULL DEFAULT 'disabled',
    ADD COLUMN runtime_process_state TEXT NOT NULL DEFAULT 'disabled';

ALTER TABLE nodes
    ADD CONSTRAINT nodes_runtime_process_mode_check CHECK (runtime_process_mode IN ('disabled', 'local')),
    ADD CONSTRAINT nodes_runtime_process_state_check CHECK (runtime_process_state IN ('disabled', 'ready', 'failed'));
