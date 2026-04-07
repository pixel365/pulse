package api

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"
)

type apiError struct {
	err    error
	status int
}

func (e apiError) Error() string {
	return e.err.Error()
}

func statusCode(err error, fallback int) int {
	var apiErr apiError
	if ok := errors.As(err, &apiErr); ok {
		return apiErr.status
	}

	return fallback
}

func errorResponse(w http.ResponseWriter, r *http.Request, status int, err error) {
	render.Status(r, status)
	render.JSON(w, r, map[string]string{
		"error": err.Error(),
	})
}
