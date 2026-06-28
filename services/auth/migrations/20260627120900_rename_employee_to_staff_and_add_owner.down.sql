UPDATE roles
SET name = 'employee',
	description = 'Standard employee access',
	updated_at = NOW()
WHERE name = 'staff'
	AND NOT EXISTS (
		SELECT 1
		FROM roles existing
		WHERE existing.name = 'employee'
	);

DELETE FROM roles
WHERE name = 'owner'
	AND NOT EXISTS (
		SELECT 1
		FROM users
		WHERE users.role_id = roles.id
	);
