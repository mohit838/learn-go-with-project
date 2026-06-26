CREATE TABLE users (
	id BIGSERIAL PRIMARY KEY,
	public_id UUID NOT NULL DEFAULT gen_random_uuid(),
	tenant_id BIGINT NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
	role_id BIGINT NOT NULL REFERENCES roles(id) ON DELETE RESTRICT,
	username VARCHAR(100) NOT NULL,
	email VARCHAR(255) NOT NULL,
	password_hash TEXT NOT NULL,
	is_active BOOLEAN NOT NULL DEFAULT TRUE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX users_public_id_unique_idx
	ON users (public_id);

CREATE UNIQUE INDEX users_tenant_username_unique_idx
	ON users (tenant_id, LOWER(username));

CREATE UNIQUE INDEX users_tenant_email_unique_idx
	ON users (tenant_id, LOWER(email));

CREATE INDEX users_tenant_id_idx
	ON users (tenant_id);

CREATE INDEX users_role_id_idx
	ON users (role_id);

CREATE INDEX users_is_active_idx
	ON users (is_active);
