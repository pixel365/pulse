package api

import (
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func Routes(r chi.Router, h *Handler) {
	public(r, h)
	internal(r, h)
}

func public(r chi.Router, h *Handler) {
	enabled, _ := strconv.ParseBool(os.Getenv("PUBLIC_API_ENABLED"))
	if !enabled {
		return
	}

	r.Route("/public", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Get("/status", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotImplemented)
			})
		})
	})
}

func internal(r chi.Router, h *Handler) {
	enabled, _ := strconv.ParseBool(os.Getenv("INTERNAL_API_ENABLED"))
	if !enabled {
		return
	}

	r.Route("/internal", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Get("/services", h.Services)
			r.Get("/services/{serviceId}/checks/state", h.ServiceCheckStates)
			r.Get("/services/{serviceId}/checks/{checkId}/executions", h.CheckExecutions)
			r.Get("/services/{serviceId}/checks/{checkId}/timeline", h.CheckTimeline)
			r.Get("/services/{serviceId}/checks/{checkId}/buckets", h.CheckBuckets)
		})
	})
}
