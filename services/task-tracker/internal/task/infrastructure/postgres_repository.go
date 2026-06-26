package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/mohit838/learn-go-with-project/internal/task/domain"
)

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(ctx context.Context, input domain.CreateTaskInput) (domain.Task, error) {
	var task domain.Task
	err := r.db.QueryRowContext(ctx, `
INSERT INTO tasks (tenant_id, tenant_slug, user_id, title, description, status, priority, image_object_key, image_url)
VALUES ($1, LOWER($2), $3, $4, $5, $6, $7, NULLIF($8, ''), NULLIF($9, ''))
RETURNING id, public_id::text, tenant_id::text, tenant_slug, user_id::text, title, COALESCE(description, ''), status, priority, COALESCE(image_object_key, ''), COALESCE(image_url, ''), is_active, created_at, updated_at`,
		input.TenantID,
		input.TenantSlug,
		input.UserID,
		input.Title,
		input.Description,
		input.Status,
		input.Priority,
		input.ImageObjectKey,
		input.ImageURL,
	).Scan(taskScanDest(&task)...)
	return task, err
}

func (r *TaskRepository) Update(ctx context.Context, input domain.UpdateTaskInput) (domain.Task, error) {
	var task domain.Task
	err := r.db.QueryRowContext(ctx, `
UPDATE tasks
SET title = $1,
	description = $2,
	status = $3,
	priority = $4,
	image_object_key = COALESCE(NULLIF($5, ''), image_object_key),
	image_url = COALESCE(NULLIF($6, ''), image_url),
	updated_at = NOW()
WHERE public_id = $7
	AND tenant_id = $8
	AND user_id = $9
	AND is_active = TRUE
RETURNING id, public_id::text, tenant_id::text, tenant_slug, user_id::text, title, COALESCE(description, ''), status, priority, COALESCE(image_object_key, ''), COALESCE(image_url, ''), is_active, created_at, updated_at`,
		input.Title,
		input.Description,
		input.Status,
		input.Priority,
		input.ImageObjectKey,
		input.ImageURL,
		input.ID,
		input.TenantID,
		input.UserID,
	).Scan(taskScanDest(&task)...)
	return task, err
}

func (r *TaskRepository) MarkInactive(ctx context.Context, tenantID, userID, id string) error {
	result, err := r.db.ExecContext(ctx, `
UPDATE tasks
SET is_active = FALSE, updated_at = NOW()
WHERE public_id = $1 AND tenant_id = $2 AND user_id = $3`, id, tenantID, userID)
	if err != nil {
		return err
	}
	return requireAffected(result)
}

func (r *TaskRepository) Delete(ctx context.Context, tenantID, userID, id string) error {
	result, err := r.db.ExecContext(ctx, `
DELETE FROM tasks
WHERE public_id = $1 AND tenant_id = $2 AND user_id = $3`, id, tenantID, userID)
	if err != nil {
		return err
	}
	return requireAffected(result)
}

func (r *TaskRepository) FindByID(ctx context.Context, tenantID, userID, id string) (domain.Task, error) {
	var task domain.Task
	err := r.db.QueryRowContext(ctx, `
SELECT id, public_id::text, tenant_id::text, tenant_slug, user_id::text, title, COALESCE(description, ''), status, priority, COALESCE(image_object_key, ''), COALESCE(image_url, ''), is_active, created_at, updated_at
FROM tasks
WHERE public_id = $1 AND tenant_id = $2 AND user_id = $3`, id, tenantID, userID).Scan(taskScanDest(&task)...)
	return task, err
}

func (r *TaskRepository) List(ctx context.Context, filter domain.ListTaskFilter) ([]domain.Task, int64, error) {
	where, args := buildTaskListWhere(filter)

	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM tasks`+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, filter.Limit, filter.Offset)
	query := `
SELECT id, public_id::text, tenant_id::text, tenant_slug, user_id::text, title, COALESCE(description, ''), status, priority, COALESCE(image_object_key, ''), COALESCE(image_url, ''), is_active, created_at, updated_at
FROM tasks` + where + `
ORDER BY created_at DESC, id DESC
LIMIT $` + fmt.Sprint(len(args)-1) + ` OFFSET $` + fmt.Sprint(len(args))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	tasks := make([]domain.Task, 0)
	for rows.Next() {
		var task domain.Task
		if err := rows.Scan(taskScanDest(&task)...); err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return tasks, total, nil
}

func buildTaskListWhere(filter domain.ListTaskFilter) (string, []any) {
	clauses := make([]string, 0)
	args := make([]any, 0)
	add := func(clause string, value any) {
		args = append(args, value)
		clauses = append(clauses, fmt.Sprintf(clause, len(args)))
	}

	add("tenant_id = $%d", filter.TenantID)
	if filter.OnlyMine {
		add("user_id = $%d", filter.UserID)
	}
	if filter.Search != "" {
		add("(title ILIKE '%%' || $%[1]d || '%%' OR description ILIKE '%%' || $%[1]d || '%%')", filter.Search)
	}
	if filter.Status != "" {
		add("status = LOWER($%d)", filter.Status)
	}
	if filter.Priority != "" {
		add("priority = LOWER($%d)", filter.Priority)
	}
	if filter.IsActive != nil {
		add("is_active = $%d", *filter.IsActive)
	}

	return " WHERE " + strings.Join(clauses, " AND "), args
}

func taskScanDest(task *domain.Task) []any {
	return []any{
		&task.ID,
		&task.PublicID,
		&task.TenantID,
		&task.TenantSlug,
		&task.UserID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.Priority,
		&task.ImageObjectKey,
		&task.ImageURL,
		&task.IsActive,
		&task.CreatedAt,
		&task.UpdatedAt,
	}
}

func requireAffected(result sql.Result) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *TaskRepository) DashboardStats(ctx context.Context) (domain.TaskDashboardStats, error) {
	var stats domain.TaskDashboardStats
	if err := r.db.QueryRowContext(ctx, `
	SELECT
		COUNT(*),
		COUNT(*) FILTER (WHERE is_active = TRUE),
		COUNT(*) FILTER (WHERE is_active = FALSE)
	FROM tasks`).Scan(&stats.Total, &stats.Active, &stats.Inactive); err != nil {
		return domain.TaskDashboardStats{}, err
	}

	statusRows, err := r.db.QueryContext(ctx, `
	SELECT status, COUNT(*)
	FROM tasks
	GROUP BY status
	ORDER BY status`)
	if err != nil {
		return domain.TaskDashboardStats{}, err
	}
	defer statusRows.Close()
	for statusRows.Next() {
		var item domain.StatusTaskCount
		if err := statusRows.Scan(&item.Status, &item.Count); err != nil {
			return domain.TaskDashboardStats{}, err
		}
		stats.ByStatus = append(stats.ByStatus, item)
	}
	if err := statusRows.Err(); err != nil {
		return domain.TaskDashboardStats{}, err
	}

	priorityRows, err := r.db.QueryContext(ctx, `
	SELECT priority, COUNT(*)
	FROM tasks
	GROUP BY priority
	ORDER BY priority`)
	if err != nil {
		return domain.TaskDashboardStats{}, err
	}
	defer priorityRows.Close()
	for priorityRows.Next() {
		var item domain.PriorityTaskCount
		if err := priorityRows.Scan(&item.Priority, &item.Count); err != nil {
			return domain.TaskDashboardStats{}, err
		}
		stats.ByPriority = append(stats.ByPriority, item)
	}
	if err := priorityRows.Err(); err != nil {
		return domain.TaskDashboardStats{}, err
	}

	userRows, err := r.db.QueryContext(ctx, `
	SELECT
		user_id::text,
		COUNT(*),
		COUNT(*) FILTER (WHERE is_active = TRUE),
		COUNT(*) FILTER (WHERE is_active = FALSE)
	FROM tasks
	GROUP BY user_id
	ORDER BY COUNT(*) DESC, user_id::text ASC`)
	if err != nil {
		return domain.TaskDashboardStats{}, err
	}
	defer userRows.Close()
	for userRows.Next() {
		var item domain.UserTaskCount
		if err := userRows.Scan(&item.UserID, &item.Total, &item.Active, &item.Inactive); err != nil {
			return domain.TaskDashboardStats{}, err
		}
		stats.ByUser = append(stats.ByUser, item)
	}
	if err := userRows.Err(); err != nil {
		return domain.TaskDashboardStats{}, err
	}

	return stats, nil
}
