package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"github.com/pixel365/pulse/internal/config"
	"github.com/pixel365/pulse/internal/model"
	executionsvc "github.com/pixel365/pulse/internal/services/execution"
	"github.com/pixel365/pulse/internal/services/state"
)

type checkExecutionResponse struct {
	StartedAt     time.Time                  `json:"started_at"`
	FinishedAt    time.Time                  `json:"finished_at"`
	CreatedAt     time.Time                  `json:"created_at"`
	Details       map[string]any             `json:"details"`
	ExecutionID   string                     `json:"execution_id"`
	CheckID       string                     `json:"check_id"`
	ServiceID     string                     `json:"service_id"`
	CheckType     config.CheckType           `json:"check_type"`
	Status        model.CheckExecutionStatus `json:"status"`
	ErrorKind     string                     `json:"error_kind"`
	ErrorMessage  string                     `json:"error_message"`
	DurationUs    int64                      `json:"duration_us"`
	AttemptsTotal int                        `json:"attempts_total"`
}

type Handler struct {
	cfgProvider config.ConfigProvider
	execution   executionsvc.Service
	state       state.StateService
}

func NewHandler(
	cfgProvider config.ConfigProvider,
	stateSvc state.StateService,
	executionSvc executionsvc.Service,
) *Handler {
	return &Handler{
		cfgProvider: cfgProvider,
		state:       stateSvc,
		execution:   executionSvc,
	}
}

func (h *Handler) Services(w http.ResponseWriter, r *http.Request) {
	render.Status(r, http.StatusOK)
	render.JSON(w, r, h.cfgProvider.Current().Services)
}

func (h *Handler) ServiceCheckStates(w http.ResponseWriter, r *http.Request) {
	serviceID := chi.URLParam(r, "serviceId")
	states, err := h.state.GetStatesByService(r.Context(), serviceID)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, states)
}

func (h *Handler) CheckExecutions(w http.ResponseWriter, r *http.Request) {
	filter, err := executionFilterFromRequest(r)
	if err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	records, err := h.execution.ListExecutions(r.Context(), filter)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, checkExecutionRecordsResponse(records))
}

func (h *Handler) CheckTimeline(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func executionFilterFromRequest(r *http.Request) (model.CheckExecutionFilter, error) {
	filter := model.CheckExecutionFilter{
		ServiceID: chi.URLParam(r, "serviceId"),
		CheckID:   chi.URLParam(r, "checkId"),
	}

	query := r.URL.Query()

	if raw := query.Get("from"); raw != "" {
		from, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return filter, err
		}
		filter.From = &from
	}

	if raw := query.Get("to"); raw != "" {
		to, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return filter, err
		}
		filter.To = &to
	}

	if raw := query.Get("limit"); raw != "" {
		limit, err := strconv.Atoi(raw)
		if err != nil {
			return filter, err
		}
		filter.Limit = limit
	}

	return filter, nil
}

func errorResponse(w http.ResponseWriter, r *http.Request, status int, err error) {
	render.Status(r, status)
	render.JSON(w, r, map[string]string{
		"error": err.Error(),
	})
}

func checkExecutionRecordsResponse(records []model.CheckExecutionRecord) []checkExecutionResponse {
	result := make([]checkExecutionResponse, 0, len(records))
	for i := range records {
		result = append(result, checkExecutionResponse{
			StartedAt:     records[i].StartedAt,
			FinishedAt:    records[i].FinishedAt,
			CreatedAt:     records[i].CreatedAt,
			Details:       records[i].Details,
			ExecutionID:   records[i].ExecutionID,
			CheckID:       records[i].CheckID,
			ServiceID:     records[i].ServiceID,
			CheckType:     records[i].CheckType,
			Status:        records[i].Status,
			ErrorKind:     string(records[i].ErrorKind),
			ErrorMessage:  records[i].ErrorMessage,
			DurationUs:    records[i].Duration.Microseconds(),
			AttemptsTotal: records[i].AttemptsTotal,
		})
	}

	return result
}
