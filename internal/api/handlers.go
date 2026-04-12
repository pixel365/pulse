package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"github.com/pixel365/pulse/internal/config"
	executionsvc "github.com/pixel365/pulse/internal/services/execution"
	"github.com/pixel365/pulse/internal/services/state"
)

type Handler struct {
	cfgProvider  config.ConfigProvider
	executionSvc executionsvc.Service
	stateSvc     state.StateService
}

func NewHandler(
	cfgProvider config.ConfigProvider,
	stateSvc state.StateService,
	executionSvc executionsvc.Service,
) *Handler {
	return &Handler{
		cfgProvider:  cfgProvider,
		stateSvc:     stateSvc,
		executionSvc: executionSvc,
	}
}

func (h *Handler) Services(w http.ResponseWriter, r *http.Request) {
	serviceStates, err := h.stateSvc.ListServiceStates(r.Context(), h.cfgProvider.Current())
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, servicesResponse(h.cfgProvider.Current(), serviceStates))
}

func (h *Handler) ServiceCheckStates(w http.ResponseWriter, r *http.Request) {
	serviceID := chi.URLParam(r, "serviceId")
	if err := h.ensureService(serviceID); err != nil {
		errorResponse(w, r, statusCode(err, http.StatusBadRequest), err)
		return
	}

	states, err := h.stateSvc.GetStatesByService(r.Context(), serviceID)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, checkStatesResponse(h.cfgProvider.Current(), serviceID, states))
}

func (h *Handler) CheckExecutions(w http.ResponseWriter, r *http.Request) {
	filter, err := executionFilterFromRequest(r)
	if err != nil {
		errorResponse(w, r, statusCode(err, http.StatusBadRequest), err)
		return
	}

	if err = h.ensureCheck(filter.ServiceID, filter.CheckID); err != nil {
		errorResponse(w, r, statusCode(err, http.StatusBadRequest), err)
		return
	}

	records, err := h.executionSvc.ListExecutions(r.Context(), filter)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, checkExecutionRecordsResponse(records))
}

func (h *Handler) CheckTimeline(w http.ResponseWriter, r *http.Request) {
	filter, err := h.timelineFilterFromRequest(r)
	if err != nil {
		errorResponse(w, r, statusCode(err, http.StatusBadRequest), err)
		return
	}

	records, err := h.executionSvc.ListExecutionTimeline(r.Context(), filter)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, checkExecutionTimelineResponseFromRecords(records))
}

func (h *Handler) CheckBuckets(w http.ResponseWriter, r *http.Request) {
	filter, err := h.bucketFilterFromRequest(r)
	if err != nil {
		errorResponse(w, r, statusCode(err, http.StatusBadRequest), err)
		return
	}

	records, err := h.executionSvc.ListExecutionBuckets(r.Context(), filter)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, checkExecutionBucketResponseFromRecords(records))
}
