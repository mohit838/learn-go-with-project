package task

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var ErrNotFound = errors.New("task not found")

type CreateRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	IsActive    *bool  `json:"isActive"`
}
type UpdateRequest = CreateRequest
type Response struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	IsActive    bool      `json:"isActive"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
type Repository struct{ DB *sql.DB }

func (r *Repository) Create(c context.Context, v CreateRequest, active bool) (Response, error) {
	return scan(r.DB.QueryRowContext(c, `INSERT INTO tasks (title,description,status,is_active) VALUES ($1,$2,$3,$4) RETURNING id,title,description,status,is_active,created_at,updated_at`, v.Title, v.Description, status(v.Status), active))
}
func (r *Repository) List(c context.Context) ([]Response, error) {
	rows, e := r.DB.QueryContext(c, `SELECT id,title,description,status,is_active,created_at,updated_at FROM tasks ORDER BY id`)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []Response
	for rows.Next() {
		v, e := scan(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (r *Repository) Get(c context.Context, id int64) (Response, error) {
	return scan(r.DB.QueryRowContext(c, `SELECT id,title,description,status,is_active,created_at,updated_at FROM tasks WHERE id=$1`, id))
}
func (r *Repository) Update(c context.Context, id int64, v UpdateRequest, active bool) (Response, error) {
	return scan(r.DB.QueryRowContext(c, `UPDATE tasks SET title=$1,description=$2,status=$3,is_active=$4,updated_at=NOW() WHERE id=$5 RETURNING id,title,description,status,is_active,created_at,updated_at`, v.Title, v.Description, status(v.Status), active, id))
}
func (r *Repository) Delete(c context.Context, id int64) error {
	x, e := r.DB.ExecContext(c, `DELETE FROM tasks WHERE id=$1`, id)
	if e != nil {
		return e
	}
	n, e := x.RowsAffected()
	if e != nil {
		return e
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

type scanner interface{ Scan(...any) error }

func scan(s scanner) (Response, error) {
	var v Response
	e := s.Scan(&v.ID, &v.Title, &v.Description, &v.Status, &v.IsActive, &v.CreatedAt, &v.UpdatedAt)
	if errors.Is(e, sql.ErrNoRows) {
		return Response{}, ErrNotFound
	}
	return v, e
}
func status(v string) string {
	if v == "" {
		return "pending"
	}
	return v
}

type Service struct{ Repo *Repository }

func (s *Service) Create(c context.Context, v CreateRequest) (Response, error) {
	if v.Title == "" {
		return Response{}, errors.New("title is required")
	}
	a := true
	if v.IsActive != nil {
		a = *v.IsActive
	}
	return s.Repo.Create(c, v, a)
}
func (s *Service) List(c context.Context) ([]Response, error)        { return s.Repo.List(c) }
func (s *Service) Get(c context.Context, id int64) (Response, error) { return s.Repo.Get(c, id) }
func (s *Service) Update(c context.Context, id int64, v UpdateRequest) (Response, error) {
	if v.Title == "" {
		return Response{}, errors.New("title is required")
	}
	a := true
	if v.IsActive != nil {
		a = *v.IsActive
	}
	return s.Repo.Update(c, id, v, a)
}
func (s *Service) Delete(c context.Context, id int64) error { return s.Repo.Delete(c, id) }
