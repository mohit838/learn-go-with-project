package taskclient

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/mohit838/learn-go-with-project/internal/constants"
	"github.com/mohit838/learn-go-with-project/internal/response"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Routes(client *Client) func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
			id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
			if err != nil || id <= 0 {
				response.Error(w, http.StatusBadRequest, constants.ErrorBadRequest, "invalid task id")
				return
			}
			value, err := client.Get(r.Context(), id)
			if errors.Is(err, ErrDisabled) {
				response.Error(w, http.StatusServiceUnavailable, constants.ErrorInternalServer, "task service client is not configured")
				return
			}
			if status.Code(err) == codes.NotFound {
				response.Error(w, http.StatusNotFound, constants.ErrorNotFound, "task not found")
				return
			}
			if err != nil {
				response.Error(w, http.StatusBadGateway, constants.ErrorInternalServer, "task service unavailable")
				return
			}
			response.JSON(w, http.StatusOK, value)
		})
	}
}
