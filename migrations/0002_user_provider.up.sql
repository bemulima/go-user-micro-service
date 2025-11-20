CREATE TABLE IF NOT EXISTS user_provider (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    provider_type text NOT NULL,
    provider_user_id text NOT NULL,
    user_id uuid NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    metadata jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (provider_type, provider_user_id)
);

CREATE INDEX IF NOT EXISTS idx_user_provider_user_id ON user_provider(user_id);
