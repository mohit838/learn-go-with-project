CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE tasks (
	id BIGSERIAL PRIMARY KEY,
	public_id UUID NOT NULL DEFAULT gen_random_uuid(),
	tenant_id UUID NOT NULL,
	tenant_slug VARCHAR(100) NOT NULL,
	user_id UUID NOT NULL,
	title VARCHAR(180) NOT NULL,
	description TEXT,
	status VARCHAR(30) NOT NULL DEFAULT 'todo',
	priority VARCHAR(30) NOT NULL DEFAULT 'normal',
	image_object_key TEXT,
	image_url TEXT,
	is_active BOOLEAN NOT NULL DEFAULT TRUE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT tasks_status_check CHECK (status IN ('todo', 'in_progress', 'done')),
	CONSTRAINT tasks_priority_check CHECK (priority IN ('low', 'normal', 'high'))
);

CREATE UNIQUE INDEX tasks_public_id_unique_idx
	ON tasks (public_id);

CREATE INDEX tasks_tenant_user_idx
	ON tasks (tenant_id, user_id);

CREATE INDEX tasks_tenant_status_idx
	ON tasks (tenant_id, status);

CREATE INDEX tasks_tenant_active_idx
	ON tasks (tenant_id, is_active);

CREATE INDEX tasks_title_search_idx
	ON tasks USING gin (to_tsvector('simple', title));
