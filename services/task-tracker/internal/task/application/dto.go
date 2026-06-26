package application

import "github.com/mohit838/learn-go-with-project/internal/utils"

type CreateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"`
	Priority    string `json:"priority,omitempty"`
}

type UpdateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"`
	Priority    string `json:"priority,omitempty"`
}

type TaskResponse struct {
	ID          string        `json:"id"`
	TenantID    string        `json:"tenant_id"`
	TenantSlug  string        `json:"tenant_slug"`
	UserID      string        `json:"user_id"`
	Title       string        `json:"title"`
	Description string        `json:"description,omitempty"`
	Status      string        `json:"status"`
	Priority    string        `json:"priority"`
	ImageURL    string        `json:"image_url,omitempty"`
	IsActive    bool          `json:"is_active"`
	CreatedAt   utils.APITime `json:"created_at"`
	UpdatedAt   utils.APITime `json:"updated_at"`
}

type TaskListQuery struct {
	Search   string
	Status   string
	Priority string
	IsActive *bool
	OnlyMine bool
	Page     int
	PerPage  int
	Offset   int
}
