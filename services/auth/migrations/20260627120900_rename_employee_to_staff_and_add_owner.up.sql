UPDATE roles
SET name = 'staff',
	description = 'Standard staff access',
	updated_at = NOW()
WHERE name = 'employee';

INSERT INTO roles (name, description)
VALUES ('owner', 'Tenant owner access')
ON CONFLICT (name) DO UPDATE
SET description = EXCLUDED.description,
	is_active = TRUE,
	updated_at = NOW();
