package api

import (
	"slices"
	"time"

	"github.com/pixel365/pulse/internal/config"
	"github.com/pixel365/pulse/internal/model"
)

type serviceResponse struct {
	ID              string                   `json:"id"`
	Name            string                   `json:"name"`
	Description     string                   `json:"description"`
	Status          model.ServiceStateStatus `json:"status"`
	TotalChecks     int                      `json:"total_checks"`
	HealthyChecks   int                      `json:"healthy_checks"`
	UnhealthyChecks int                      `json:"unhealthy_checks"`
	UnknownChecks   int                      `json:"unknown_checks"`
}

type checkStateResponse struct {
	UpdatedAt            time.Time                  `json:"updated_at"`
	LastFailureAt        *time.Time                 `json:"last_failure_at"`
	LastSuccessAt        *time.Time                 `json:"last_success_at"`
	LastDetails          map[string]any             `json:"last_details"`
	LastErrorMessage     string                     `json:"last_error_message"`
	LastExecutionID      string                     `json:"last_execution_id"`
	CheckID              string                     `json:"check_id"`
	LastErrorKind        string                     `json:"last_error_kind"`
	LastStatus           model.CheckExecutionStatus `json:"last_status"`
	Status               model.CheckStateStatus     `json:"status"`
	CheckType            config.CheckType           `json:"check_type"`
	ServiceID            string                     `json:"service_id"`
	StatusImpact         config.StatusImpact        `json:"status_impact"`
	LastDurationUs       int64                      `json:"last_duration_us"`
	ConsecutiveSuccesses int                        `json:"consecutive_successes"`
	ConsecutiveFailures  int                        `json:"consecutive_failures"`
}

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

type checkExecutionTimelineResponse struct {
	BucketStart         time.Time                   `json:"bucket_start"`
	BucketEnd           time.Time                   `json:"bucket_end"`
	LastObservedAt      *time.Time                  `json:"last_observed_at"`
	LastExecutionStatus *model.CheckExecutionStatus `json:"last_execution_status"`
	State               model.CheckStateStatus      `json:"state"`
}

type checkExecutionBucketResponse struct {
	BucketStart   time.Time `json:"bucket_start"`
	Total         int       `json:"total"`
	SuccessCount  int       `json:"success_count"`
	FailureCount  int       `json:"failure_count"`
	AvgDurationUs int64     `json:"avg_duration_us"`
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

func checkExecutionTimelineResponseFromRecords(
	records []model.CheckExecutionTimelineRecord,
) []checkExecutionTimelineResponse {
	result := make([]checkExecutionTimelineResponse, 0, len(records))
	for i := range records {
		result = append(result, checkExecutionTimelineResponse{
			BucketStart:         records[i].BucketStart,
			BucketEnd:           records[i].BucketEnd,
			LastObservedAt:      records[i].LastObservedAt,
			LastExecutionStatus: records[i].LastExecutionStatus,
			State:               records[i].State,
		})
	}

	return result
}

func checkExecutionBucketResponseFromRecords(
	records []model.CheckExecutionBucketRecord,
) []checkExecutionBucketResponse {
	result := make([]checkExecutionBucketResponse, 0, len(records))
	for i := range records {
		result = append(result, checkExecutionBucketResponse{
			BucketStart:   records[i].BucketStart,
			Total:         records[i].Total,
			SuccessCount:  records[i].SuccessCount,
			FailureCount:  records[i].FailureCount,
			AvgDurationUs: records[i].AvgDurationUs,
		})
	}

	return result
}

func servicesResponse(
	cfg *config.Config,
	serviceStates []model.ServiceState,
) []serviceResponse {
	stateByServiceID := make(map[string]model.ServiceState, len(serviceStates))
	for _, state := range serviceStates {
		stateByServiceID[state.ServiceID] = state
	}

	serviceIDs := make([]string, 0, len(cfg.Services))
	for serviceID := range cfg.Services {
		serviceIDs = append(serviceIDs, serviceID)
	}
	slices.Sort(serviceIDs)

	result := make([]serviceResponse, 0, len(serviceIDs))
	for _, serviceID := range serviceIDs {
		service := cfg.Services[serviceID]
		state := stateByServiceID[serviceID]

		result = append(result, serviceResponse{
			ID:              service.ID,
			Name:            service.Name,
			Description:     service.Description,
			Status:          state.Status,
			TotalChecks:     state.TotalChecks,
			HealthyChecks:   state.HealthyChecks,
			UnhealthyChecks: state.UnhealthyChecks,
			UnknownChecks:   state.UnknownChecks,
		})
	}

	return result
}

func checkStatesResponse(
	cfg *config.Config,
	serviceID string,
	states []model.CheckState,
) []checkStateResponse {
	fieldsByCheckID := checkFieldsByService(cfg, serviceID)

	result := make([]checkStateResponse, 0, len(states))
	for k := range states {
		fields := fieldsByCheckID[states[k].CheckID]
		result = append(result, checkStateResponse{
			UpdatedAt:            states[k].UpdatedAt,
			LastFailureAt:        states[k].LastFailureAt,
			LastSuccessAt:        states[k].LastSuccessAt,
			LastDetails:          states[k].LastDetails,
			LastStatus:           states[k].LastStatus,
			LastExecutionID:      states[k].LastExecutionID,
			CheckID:              states[k].CheckID,
			LastErrorKind:        string(states[k].LastErrorKind),
			LastErrorMessage:     states[k].LastErrorMessage,
			Status:               states[k].Status,
			CheckType:            states[k].CheckType,
			ServiceID:            states[k].ServiceID,
			LastDurationUs:       states[k].LastDuration.Microseconds(),
			ConsecutiveSuccesses: states[k].ConsecutiveSuccesses,
			ConsecutiveFailures:  states[k].ConsecutiveFailures,
			StatusImpact:         fields.StatusImpact,
		})
	}

	return result
}

func checkFieldsByService(cfg *config.Config, serviceID string) map[string]config.CheckFields {
	result := map[string]config.CheckFields{}

	appendCheckFieldsByService(result, serviceID, cfg.HttpChecks)
	appendCheckFieldsByService(result, serviceID, cfg.TCPChecks)
	appendCheckFieldsByService(result, serviceID, cfg.GRPCChecks)
	appendCheckFieldsByService(result, serviceID, cfg.DNSChecks)
	appendCheckFieldsByService(result, serviceID, cfg.TLSChecks)

	return result
}

func appendCheckFieldsByService[T any](
	dst map[string]config.CheckFields,
	serviceID string,
	checks map[string]config.TypedCheck[T],
) {
	for _, check := range checks {
		if check.Service != serviceID {
			continue
		}

		dst[check.ID] = check.CheckFields
	}
}
