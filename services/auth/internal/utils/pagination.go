package utils

import (
	"math"
	"net/http"
	"strconv"

	"github.com/mohit838/learn-go-with-project/internal/constants"
)

type Pagination struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
	Offset  int `json:"-"`
}

type PaginationMeta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

type PaginatedResponse[T any] struct {
	Items []T            `json:"items"`
	Meta  PaginationMeta `json:"meta"`
}

func ParsePagination(r *http.Request) Pagination {
	page := parsePositiveInt(r.URL.Query().Get("page"), constants.DefaultPage)
	perPage := parsePositiveInt(r.URL.Query().Get("per_page"), constants.DefaultPerPage)
	if perPage > constants.DefaultMaxPerPage {
		perPage = constants.DefaultMaxPerPage
	}

	return Pagination{
		Page:    page,
		PerPage: perPage,
		Offset:  (page - 1) * perPage,
	}
}

func NewPaginationMeta(p Pagination, total int64) PaginationMeta {
	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(p.PerPage)))
	}

	return PaginationMeta{
		Page:       p.Page,
		PerPage:    p.PerPage,
		Total:      total,
		TotalPages: totalPages,
	}
}

func NewPaginatedResponse[T any](items []T, p Pagination, total int64) PaginatedResponse[T] {
	if items == nil {
		items = []T{}
	}

	return PaginatedResponse[T]{
		Items: items,
		Meta:  NewPaginationMeta(p, total),
	}
}

func parsePositiveInt(value string, fallback int) int {
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return fallback
	}
	return parsed
}
