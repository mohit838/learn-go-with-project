CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE tenants (
	id BIGSERIAL PRIMARY KEY,
	public_id UUID NOT NULL DEFAULT gen_random_uuid(),
	name VARCHAR(150) NOT NULL,
	slug VARCHAR(100) NOT NULL UNIQUE,
	is_active BOOLEAN NOT NULL DEFAULT TRUE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX tenants_public_id_unique_idx
	ON tenants (public_id);

CREATE INDEX tenants_is_active_idx
	ON tenants (is_active);
