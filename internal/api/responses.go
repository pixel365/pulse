package api

import (
	"time"

	"github.com/pixel365/pulse/internal/config"
	"github.com/pixel365/pulse/internal/model"
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
