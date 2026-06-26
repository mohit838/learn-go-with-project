package application

import (
	"context"
	"errors"
	"strings"

	"github.com/mohit838/learn-go-with-project/internal/constants"
	"github.com/mohit838/learn-go-with-project/internal/task/domain"
)

var ErrForbidden = errors.New("forbidden")

type AuthStatsProvider interface {
	UserStats(ctx context.Context) (UserStatsResponse, error)
}

type DashboardService struct {
	repo domain.Repository
	auth AuthStatsProvider
}

func NewDashboardService(repo domain.Repository, auth AuthStatsProvider) *DashboardService {
	return &DashboardService{repo: repo, auth: auth}
}

func (s *DashboardService) Summary(ctx context.Context, user domain.UserContext) (DashboardResponse, error) {
	if !isSuperadmin(user.Role) {
		return DashboardResponse{}, ErrForbidden
	}

	taskStats, err := s.repo.DashboardStats(ctx)
	if err != nil {
		return DashboardResponse{}, err
	}
	userStats, err := s.auth.UserStats(ctx)
	if err != nil {
		return DashboardResponse{}, err
	}

	return DashboardResponse{
		Users: userStats,
		Tasks: taskStatsResponse(taskStats),
	}, nil
}

type DashboardResponse struct {
	Users UserStatsResponse `json:"users"`
	Tasks TaskStatsResponse `json:"tasks"`
}

type UserStatsResponse struct {
	Total    int64                     `json:"total"`
	Active   int64                     `json:"active"`
	Inactive int64                     `json:"inactive"`
	ByRole   []RoleUserCountResponse   `json:"by_role"`
	ByTenant []TenantUserCountResponse `json:"by_tenant"`
}

type RoleUserCountResponse struct {
	Role  string `json:"role"`
	Count int64  `json:"count"`
}

type TenantUserCountResponse struct {
	TenantID   string `json:"tenant_id"`
	TenantName string `json:"tenant_name"`
	TenantSlug string `json:"tenant_slug"`
	Count      int64  `json:"count"`
}

type TaskStatsResponse struct {
	Total      int64                       `json:"total"`
	Active     int64                       `json:"active"`
	Inactive   int64                       `json:"inactive"`
	ByStatus   []StatusTaskCountResponse   `json:"by_status"`
	ByPriority []PriorityTaskCountResponse `json:"by_priority"`
	ByUser     []UserTaskCountResponse     `json:"by_user"`
}

type StatusTaskCountResponse struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

type PriorityTaskCountResponse struct {
	Priority string `json:"priority"`
	Count    int64  `json:"count"`
}

type UserTaskCountResponse struct {
	UserID   string `json:"user_id"`
	Total    int64  `json:"total"`
	Active   int64  `json:"active"`
	Inactive int64  `json:"inactive"`
}

func taskStatsResponse(stats domain.TaskDashboardStats) TaskStatsResponse {
	res := TaskStatsResponse{
		Total:      stats.Total,
		Active:     stats.Active,
		Inactive:   stats.Inactive,
		ByStatus:   make([]StatusTaskCountResponse, 0, len(stats.ByStatus)),
		ByPriority: make([]PriorityTaskCountResponse, 0, len(stats.ByPriority)),
		ByUser:     make([]UserTaskCountResponse, 0, len(stats.ByUser)),
	}
	for _, item := range stats.ByStatus {
		res.ByStatus = append(res.ByStatus, StatusTaskCountResponse{
			Status: item.Status,
			Count:  item.Count,
		})
	}
	for _, item := range stats.ByPriority {
		res.ByPriority = append(res.ByPriority, PriorityTaskCountResponse{
			Priority: item.Priority,
			Count:    item.Count,
		})
	}
	for _, item := range stats.ByUser {
		res.ByUser = append(res.ByUser, UserTaskCountResponse{
			UserID:   item.UserID,
			Total:    item.Total,
			Active:   item.Active,
			Inactive: item.Inactive,
		})
	}
	return res
}

func isSuperadmin(role string) bool {
	return strings.EqualFold(strings.TrimSpace(role), constants.DefaultRoleSuperadmin)
}
