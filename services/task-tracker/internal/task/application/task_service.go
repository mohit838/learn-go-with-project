package application

import (
	"context"
	"errors"
	"strings"

	"github.com/mohit838/learn-go-with-project/internal/constants"
	"github.com/mohit838/learn-go-with-project/internal/task/domain"
	"github.com/mohit838/learn-go-with-project/internal/utils"
)

var ErrInvalidInput = errors.New("invalid input")

type TaskService struct {
	repo   domain.Repository
	images domain.ImageStorage
}

func NewTaskService(repo domain.Repository, images domain.ImageStorage) *TaskService {
	return &TaskService{repo: repo, images: images}
}

func (s *TaskService) Create(ctx context.Context, user domain.UserContext, req CreateTaskRequest, upload *domain.ImageUpload) (TaskResponse, error) {
	req = normalizeCreate(req)
	if user.UserID == "" || user.TenantID == "" || req.Title == "" {
		return TaskResponse{}, ErrInvalidInput
	}
	if !validTaskStatus(req.Status) || !validTaskPriority(req.Priority) {
		return TaskResponse{}, ErrInvalidInput
	}

	objectKey, imageURL, err := s.uploadImage(ctx, user, upload)
	if err != nil {
		return TaskResponse{}, err
	}

	task, err := s.repo.Create(ctx, domain.CreateTaskInput{
		TenantID:       user.TenantID,
		TenantSlug:     user.TenantSlug,
		UserID:         user.UserID,
		Title:          req.Title,
		Description:    req.Description,
		Status:         req.Status,
		Priority:       req.Priority,
		ImageObjectKey: objectKey,
		ImageURL:       imageURL,
	})
	if err != nil {
		return TaskResponse{}, err
	}

	return taskResponse(task), nil
}

func (s *TaskService) Update(ctx context.Context, user domain.UserContext, id string, req UpdateTaskRequest, upload *domain.ImageUpload) (TaskResponse, error) {
	req = normalizeUpdate(req)
	id = strings.TrimSpace(id)
	if user.UserID == "" || user.TenantID == "" || id == "" || req.Title == "" {
		return TaskResponse{}, ErrInvalidInput
	}
	if !validTaskStatus(req.Status) || !validTaskPriority(req.Priority) {
		return TaskResponse{}, ErrInvalidInput
	}

	objectKey, imageURL, err := s.uploadImage(ctx, user, upload)
	if err != nil {
		return TaskResponse{}, err
	}

	task, err := s.repo.Update(ctx, domain.UpdateTaskInput{
		ID:             id,
		TenantID:       user.TenantID,
		UserID:         user.UserID,
		Title:          req.Title,
		Description:    req.Description,
		Status:         req.Status,
		Priority:       req.Priority,
		ImageObjectKey: objectKey,
		ImageURL:       imageURL,
	})
	if err != nil {
		return TaskResponse{}, err
	}

	return taskResponse(task), nil
}

func (s *TaskService) MarkInactive(ctx context.Context, user domain.UserContext, id string) error {
	if user.UserID == "" || user.TenantID == "" || strings.TrimSpace(id) == "" {
		return ErrInvalidInput
	}
	return s.repo.MarkInactive(ctx, user.TenantID, user.UserID, strings.TrimSpace(id))
}

func (s *TaskService) Delete(ctx context.Context, user domain.UserContext, id string) error {
	if user.UserID == "" || user.TenantID == "" || strings.TrimSpace(id) == "" {
		return ErrInvalidInput
	}
	return s.repo.Delete(ctx, user.TenantID, user.UserID, strings.TrimSpace(id))
}

func (s *TaskService) FindByID(ctx context.Context, user domain.UserContext, id string) (TaskResponse, error) {
	if user.UserID == "" || user.TenantID == "" || strings.TrimSpace(id) == "" {
		return TaskResponse{}, ErrInvalidInput
	}
	userID := user.UserID
	if canListTenantTasks(user.Role) {
		userID = ""
	}
	task, err := s.repo.FindByID(ctx, user.TenantID, userID, strings.TrimSpace(id))
	if err != nil {
		return TaskResponse{}, err
	}
	return taskResponse(task), nil
}

func (s *TaskService) List(ctx context.Context, user domain.UserContext, query TaskListQuery) (utils.PaginatedResponse[TaskResponse], error) {
	if user.UserID == "" || user.TenantID == "" {
		return utils.PaginatedResponse[TaskResponse]{}, ErrInvalidInput
	}
	if !canListTenantTasks(user.Role) {
		query.OnlyMine = true
	}

	tasks, total, err := s.repo.List(ctx, domain.ListTaskFilter{
		TenantID: user.TenantID,
		UserID:   user.UserID,
		Search:   strings.TrimSpace(query.Search),
		Status:   strings.TrimSpace(query.Status),
		Priority: strings.TrimSpace(query.Priority),
		IsActive: query.IsActive,
		OnlyMine: query.OnlyMine,
		Limit:    query.PerPage,
		Offset:   query.Offset,
	})
	if err != nil {
		return utils.PaginatedResponse[TaskResponse]{}, err
	}

	items := make([]TaskResponse, 0, len(tasks))
	for _, task := range tasks {
		items = append(items, taskResponse(task))
	}

	return utils.NewPaginatedResponse(items, utils.Pagination{
		Page:    query.Page,
		PerPage: query.PerPage,
		Offset:  query.Offset,
	}, total), nil
}

func canListTenantTasks(role string) bool {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case constants.DefaultRoleSuperadmin, constants.DefaultRoleAdmin, constants.DefaultRoleOwner:
		return true
	default:
		return false
	}
}

func validTaskStatus(status string) bool {
	switch status {
	case "todo", "in_progress", "done":
		return true
	default:
		return false
	}
}

func validTaskPriority(priority string) bool {
	switch priority {
	case "low", "normal", "high":
		return true
	default:
		return false
	}
}

func (s *TaskService) uploadImage(ctx context.Context, user domain.UserContext, upload *domain.ImageUpload) (string, string, error) {
	if upload == nil {
		return "", "", nil
	}
	return s.images.UploadTaskImage(ctx, user.TenantID, user.UserID, *upload)
}

func normalizeCreate(req CreateTaskRequest) CreateTaskRequest {
	return CreateTaskRequest(normalizeTaskFields(req.Title, req.Description, req.Status, req.Priority))
}

func normalizeUpdate(req UpdateTaskRequest) UpdateTaskRequest {
	values := normalizeTaskFields(req.Title, req.Description, req.Status, req.Priority)
	return UpdateTaskRequest(values)
}

func normalizeTaskFields(title, description, status, priority string) struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"`
	Priority    string `json:"priority,omitempty"`
} {
	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)
	status = strings.TrimSpace(status)
	priority = strings.TrimSpace(priority)
	if status == "" {
		status = "todo"
	}
	if priority == "" {
		priority = "normal"
	}
	return struct {
		Title       string `json:"title"`
		Description string `json:"description,omitempty"`
		Status      string `json:"status,omitempty"`
		Priority    string `json:"priority,omitempty"`
	}{title, description, status, priority}
}

func taskResponse(task domain.Task) TaskResponse {
	return TaskResponse{
		ID:          task.PublicID,
		TenantID:    task.TenantID,
		TenantSlug:  task.TenantSlug,
		UserID:      task.UserID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		Priority:    task.Priority,
		ImageURL:    task.ImageURL,
		IsActive:    task.IsActive,
		CreatedAt:   utils.NewAPITime(task.CreatedAt),
		UpdatedAt:   utils.NewAPITime(task.UpdatedAt),
	}
}
