package api

import "github.com/go-chi/chi/v5"

func Routes(r chi.Router, h *Handler) {
	r.Route("/v1", func(r chi.Router) {
		r.Get("/services", h.Services)
		r.Get("/services/{serviceId}/checks/state", h.ServiceCheckStates)
		r.Get("/services/{serviceId}/checks/{checkId}/executions", h.CheckExecutions)
		r.Get("/services/{serviceId}/checks/{checkId}/timeline", h.CheckTimeline)
		r.Get("/services/{serviceId}/checks/{checkId}/buckets", h.CheckBuckets)
	})
}
