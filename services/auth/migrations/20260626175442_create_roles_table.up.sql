CREATE TABLE roles (
	id BIGSERIAL PRIMARY KEY,
	name VARCHAR(50) NOT NULL UNIQUE,
	description TEXT,
	is_active BOOLEAN NOT NULL DEFAULT TRUE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX roles_is_active_idx
	ON roles (is_active);

INSERT INTO roles (name, description) VALUES
	('superadmin', 'Full platform access'),
	('admin', 'Tenant administration access'),
	('employee', 'Standard employee access'),
	('guest', 'Limited guest access');
