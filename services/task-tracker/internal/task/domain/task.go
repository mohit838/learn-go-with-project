package domain

import (
	"context"
	"io"
	"time"
)

type UserContext struct {
	UserID     string
	TenantID   string
	TenantSlug string
	Role       string
}

type Task struct {
	ID             int64
	PublicID       string
	TenantID       string
	TenantSlug     string
	UserID         string
	Title          string
	Description    string
	Status         string
	Priority       string
	ImageObjectKey string
	ImageURL       string
	IsActive       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type CreateTaskInput struct {
	TenantID       string
	TenantSlug     string
	UserID         string
	Title          string
	Description    string
	Status         string
	Priority       string
	ImageObjectKey string
	ImageURL       string
}

type UpdateTaskInput struct {
	ID             string
	TenantID       string
	UserID         string
	Title          string
	Description    string
	Status         string
	Priority       string
	ImageObjectKey string
	ImageURL       string
}

type ListTaskFilter struct {
	TenantID string
	UserID   string
	Search   string
	Status   string
	Priority string
	IsActive *bool
	OnlyMine bool
	Limit    int
	Offset   int
}

type Repository interface {
	Create(ctx context.Context, input CreateTaskInput) (Task, error)
	Update(ctx context.Context, input UpdateTaskInput) (Task, error)
	MarkInactive(ctx context.Context, tenantID, userID, id string) error
	Delete(ctx context.Context, tenantID, userID, id string) error
	FindByID(ctx context.Context, tenantID, userID, id string) (Task, error)
	List(ctx context.Context, filter ListTaskFilter) ([]Task, int64, error)
}

type ImageUpload struct {
	FileName    string
	ContentType string
	Size        int64
	Reader      io.Reader
	Closer      io.Closer
}

type ImageStorage interface {
	UploadTaskImage(ctx context.Context, tenantID, userID string, upload ImageUpload) (objectKey string, publicURL string, err error)
}
