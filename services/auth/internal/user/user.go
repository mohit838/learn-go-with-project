package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/mohit838/learn-go-with-project/internal/constants"
	"github.com/mohit838/learn-go-with-project/internal/response"
	"golang.org/x/crypto/bcrypt"
)

var ErrNotFound = errors.New("user not found")

type CreateRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	IsActive *bool  `json:"isActive"`
}

type UpdateRequest = CreateRequest

type Response struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type entity struct {
	Response
	passwordHash string
}

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Create(ctx context.Context, username, passwordHash string, active bool) (entity, error) {
	return scan(r.db.QueryRowContext(ctx, `INSERT INTO users (username, password_hash, is_active) VALUES ($1, $2, $3) RETURNING id, username, password_hash, is_active, created_at, updated_at`, username, passwordHash, active))
}

func (r *Repository) List(ctx context.Context) ([]entity, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, username, password_hash, is_active, created_at, updated_at FROM users ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []entity
	for rows.Next() {
		user, err := scan(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (r *Repository) Get(ctx context.Context, id int64) (entity, error) {
	return scan(r.db.QueryRowContext(ctx, `SELECT id, username, password_hash, is_active, created_at, updated_at FROM users WHERE id = $1`, id))
}

func (r *Repository) Update(ctx context.Context, id int64, username, passwordHash string, active bool) (entity, error) {
	return scan(r.db.QueryRowContext(ctx, `UPDATE users SET username = $1, password_hash = $2, is_active = $3, updated_at = NOW() WHERE id = $4 RETURNING id, username, password_hash, is_active, created_at, updated_at`, username, passwordHash, active, id))
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

type scanner interface{ Scan(...any) error }

func scan(row scanner) (entity, error) {
	var user entity
	err := row.Scan(&user.ID, &user.Username, &user.passwordHash, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return entity{}, ErrNotFound
	}
	return user, err
}

type Service struct{ repo *Repository }

func NewService(repo *Repository) *Service { return &Service{repo: repo} }
func (s *Service) Create(ctx context.Context, input CreateRequest) (Response, error) {
	if err := validate(input); err != nil {
		return Response{}, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return Response{}, err
	}
	active := true
	if input.IsActive != nil {
		active = *input.IsActive
	}
	user, err := s.repo.Create(ctx, input.Username, string(hash), active)
	return user.Response, err
}
func (s *Service) List(ctx context.Context) ([]Response, error) {
	users, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Response, len(users))
	for i := range users {
		out[i] = users[i].Response
	}
	return out, nil
}
func (s *Service) Get(ctx context.Context, id int64) (Response, error) {
	user, err := s.repo.Get(ctx, id)
	return user.Response, err
}
func (s *Service) Update(ctx context.Context, id int64, input UpdateRequest) (Response, error) {
	if err := validate(input); err != nil {
		return Response{}, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return Response{}, err
	}
	active := true
	if input.IsActive != nil {
		active = *input.IsActive
	}
	user, err := s.repo.Update(ctx, id, input.Username, string(hash), active)
	return user.Response, err
}
func (s *Service) Delete(ctx context.Context, id int64) error { return s.repo.Delete(ctx, id) }
func validate(input CreateRequest) error {
	if strings.TrimSpace(input.Username) == "" || len(input.Password) < 8 {
		return errors.New("username is required and password must be at least 8 characters")
	}
	return nil
}

type Handler struct{ service *Service }

func NewHandler(db *sql.DB) *Handler { return &Handler{service: NewService(NewRepository(db))} }
func (h *Handler) Routes(r chi.Router) {
	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Get("/{id}", h.get)
	r.Put("/{id}", h.update)
	r.Delete("/{id}", h.delete)
}
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var input CreateRequest
	if !decode(r, &input) {
		response.Error(w, http.StatusBadRequest, constants.ErrorBadRequest, "invalid request body")
		return
	}
	value, err := h.service.Create(r.Context(), input)
	h.write(w, http.StatusCreated, value, err)
}
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	value, err := h.service.List(r.Context())
	h.write(w, http.StatusOK, value, err)
}
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, ok := idFrom(r)
	if !ok {
		response.Error(w, http.StatusBadRequest, constants.ErrorBadRequest, "invalid user id")
		return
	}
	value, err := h.service.Get(r.Context(), id)
	h.write(w, http.StatusOK, value, err)
}
func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, ok := idFrom(r)
	var input UpdateRequest
	if !ok || !decode(r, &input) {
		response.Error(w, http.StatusBadRequest, constants.ErrorBadRequest, "invalid request")
		return
	}
	value, err := h.service.Update(r.Context(), id, input)
	h.write(w, http.StatusOK, value, err)
}
func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, ok := idFrom(r)
	if !ok {
		response.Error(w, http.StatusBadRequest, constants.ErrorBadRequest, "invalid user id")
		return
	}
	h.write(w, http.StatusOK, map[string]bool{"deleted": true}, h.service.Delete(r.Context(), id))
}
func (h *Handler) write(w http.ResponseWriter, status int, value any, err error) {
	if err == nil {
		response.JSON(w, status, value)
		return
	}
	if errors.Is(err, ErrNotFound) {
		response.Error(w, http.StatusNotFound, constants.ErrorNotFound, err.Error())
		return
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		response.Error(w, http.StatusConflict, constants.ErrorConflict, "username already exists")
		return
	}
	if strings.Contains(err.Error(), "username is required") {
		response.Error(w, http.StatusBadRequest, constants.ErrorBadRequest, err.Error())
		return
	}
	response.Error(w, http.StatusInternalServerError, constants.ErrorInternalServer, "internal server error")
}
func decode(r *http.Request, value any) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(value) == nil
}
func idFrom(r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	return id, err == nil && id > 0
}
