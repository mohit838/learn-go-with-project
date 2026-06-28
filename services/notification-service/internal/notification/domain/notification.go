package domain

import "time"

type Status string

const (
	StatusQueued    Status = "queued"
	StatusSending   Status = "sending"
	StatusDelivered Status = "delivered"
	StatusFailed    Status = "failed"
)

type Notification struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	TenantID    string    `json:"tenant_id"`
	Recipient   string    `json:"recipient"`
	Message     string    `json:"message"`
	Status      Status    `json:"status"`
	WorkerID    int       `json:"worker_id,omitempty"`
	Error       string    `json:"error,omitempty"`
	SubmittedAt time.Time `json:"submitted_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UserContext struct {
	UserID     string
	TenantID   string
	TenantSlug string
	Role       string
}
