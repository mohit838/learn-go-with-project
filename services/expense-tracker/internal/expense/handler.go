package expense

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/mohit838/learn-go-with-project/internal/constants"
	"github.com/mohit838/learn-go-with-project/internal/response"
	"net/http"
	"strconv"
)

type Handler struct{ Service *Service }

func NewHandler(db *sql.DB) *Handler { return &Handler{&Service{&Repository{db}}} }
func (h *Handler) Routes(r chi.Router) {
	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Get("/{id}", h.get)
	r.Put("/{id}", h.update)
	r.Delete("/{id}", h.delete)
}
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var v CreateRequest
	if !decode(r, &v) {
		bad(w, "invalid request body")
		return
	}
	x, e := h.Service.Create(r.Context(), v)
	write(w, 201, x, e)
}
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	x, e := h.Service.List(r.Context())
	write(w, 200, x, e)
}
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, ok := id(r)
	if !ok {
		bad(w, "invalid expense id")
		return
	}
	x, e := h.Service.Get(r.Context(), id)
	write(w, 200, x, e)
}
func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, ok := id(r)
	var v UpdateRequest
	if !ok || !decode(r, &v) {
		bad(w, "invalid request")
		return
	}
	x, e := h.Service.Update(r.Context(), id, v)
	write(w, 200, x, e)
}
func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, ok := id(r)
	if !ok {
		bad(w, "invalid expense id")
		return
	}
	write(w, 200, map[string]bool{"deleted": true}, h.Service.Delete(r.Context(), id))
}
func write(w http.ResponseWriter, s int, v any, e error) {
	if e == nil {
		response.JSON(w, s, v)
	} else if errors.Is(e, ErrNotFound) {
		response.Error(w, 404, constants.ErrorNotFound, e.Error())
	} else if e.Error() == "title, category, and a non-negative amount are required" {
		bad(w, e.Error())
	} else {
		response.Error(w, 500, constants.ErrorInternalServer, "internal server error")
	}
}
func bad(w http.ResponseWriter, m string) { response.Error(w, 400, constants.ErrorBadRequest, m) }
func decode(r *http.Request, v any) bool  { return json.NewDecoder(r.Body).Decode(v) == nil }
func id(r *http.Request) (int64, bool) {
	v, e := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	return v, e == nil && v > 0
}
