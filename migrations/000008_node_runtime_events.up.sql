ALTER TABLE nodes
    ADD COLUMN runtime_events JSONB NOT NULL DEFAULT '[]'::jsonb;
