package expense

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var ErrNotFound = errors.New("expense not found")

type CreateRequest struct {
	Title    string  `json:"title"`
	Amount   float64 `json:"amount"`
	Category string  `json:"category"`
	IsActive *bool   `json:"isActive"`
}
type UpdateRequest = CreateRequest
type Response struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Amount    float64   `json:"amount"`
	Category  string    `json:"category"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
type Repository struct{ DB *sql.DB }

func (r *Repository) Create(c context.Context, v CreateRequest, a bool) (Response, error) {
	return scan(r.DB.QueryRowContext(c, `INSERT INTO expenses (title,amount,category,is_active) VALUES ($1,$2,$3,$4) RETURNING id,title,amount,category,is_active,created_at,updated_at`, v.Title, v.Amount, v.Category, a))
}
func (r *Repository) List(c context.Context) ([]Response, error) {
	rows, e := r.DB.QueryContext(c, `SELECT id,title,amount,category,is_active,created_at,updated_at FROM expenses ORDER BY id`)
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
	return scan(r.DB.QueryRowContext(c, `SELECT id,title,amount,category,is_active,created_at,updated_at FROM expenses WHERE id=$1`, id))
}
func (r *Repository) Update(c context.Context, id int64, v UpdateRequest, a bool) (Response, error) {
	return scan(r.DB.QueryRowContext(c, `UPDATE expenses SET title=$1,amount=$2,category=$3,is_active=$4,updated_at=NOW() WHERE id=$5 RETURNING id,title,amount,category,is_active,created_at,updated_at`, v.Title, v.Amount, v.Category, a, id))
}
func (r *Repository) Delete(c context.Context, id int64) error {
	x, e := r.DB.ExecContext(c, `DELETE FROM expenses WHERE id=$1`, id)
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
	e := s.Scan(&v.ID, &v.Title, &v.Amount, &v.Category, &v.IsActive, &v.CreatedAt, &v.UpdatedAt)
	if errors.Is(e, sql.ErrNoRows) {
		return Response{}, ErrNotFound
	}
	return v, e
}

type Service struct{ Repo *Repository }

func (s *Service) Create(c context.Context, v CreateRequest) (Response, error) {
	if v.Title == "" || v.Category == "" || v.Amount < 0 {
		return Response{}, errors.New("title, category, and a non-negative amount are required")
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
	if v.Title == "" || v.Category == "" || v.Amount < 0 {
		return Response{}, errors.New("title, category, and a non-negative amount are required")
	}
	a := true
	if v.IsActive != nil {
		a = *v.IsActive
	}
	return s.Repo.Update(c, id, v, a)
}
func (s *Service) Delete(c context.Context, id int64) error { return s.Repo.Delete(c, id) }
