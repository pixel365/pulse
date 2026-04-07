package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/pixel365/pulse/internal/model"
)

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

func (h *Handler) timelineFilterFromRequest(
	r *http.Request,
) (model.CheckExecutionTimelineFilter, error) {
	filter := model.CheckExecutionTimelineFilter{
		ServiceID: chi.URLParam(r, "serviceId"),
		CheckID:   chi.URLParam(r, "checkId"),
	}

	query := r.URL.Query()

	if raw := query.Get("from"); raw != "" {
		from, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return filter, err
		}
		filter.From = from
	}

	if raw := query.Get("to"); raw != "" {
		to, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return filter, err
		}
		filter.To = to
	}

	filter.Bucket = model.CheckExecutionBucket(query.Get("bucket"))

	fields, err := h.checkFields(filter.ServiceID, filter.CheckID)
	if err != nil {
		return filter, err
	}
	filter.Interval = fields.Interval

	if !isAllowedBucket(filter.Bucket, fields.AllowedBuckets) {
		return filter, fmt.Errorf(
			"bucket %q is not allowed for check %s/%s; allowed buckets: %v",
			filter.Bucket,
			filter.ServiceID,
			filter.CheckID,
			allowedBuckets(fields.AllowedBuckets),
		)
	}

	return filter, filter.Validate()
}

func (h *Handler) bucketFilterFromRequest(
	r *http.Request,
) (model.CheckExecutionAggregateFilter, error) {
	filter := model.CheckExecutionAggregateFilter{
		CheckExecutionFilter: model.CheckExecutionFilter{
			ServiceID: chi.URLParam(r, "serviceId"),
			CheckID:   chi.URLParam(r, "checkId"),
		},
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

	filter.Bucket = model.CheckExecutionBucket(query.Get("bucket"))

	fields, err := h.checkFields(filter.ServiceID, filter.CheckID)
	if err != nil {
		return filter, err
	}

	if !isAllowedBucket(filter.Bucket, fields.AllowedBuckets) {
		return filter, fmt.Errorf(
			"bucket %q is not allowed for check %s/%s; allowed buckets: %v",
			filter.Bucket,
			filter.ServiceID,
			filter.CheckID,
			allowedBuckets(fields.AllowedBuckets),
		)
	}

	return filter, validateBucketFilter(filter)
}

func allowedBuckets(allowed []string) []string {
	if len(allowed) > 0 {
		return allowed
	}

	return []string{
		string(model.CheckExecutionBucketMinute),
		string(model.CheckExecutionBucketHour),
	}
}

func isAllowedBucket(bucket model.CheckExecutionBucket, allowed []string) bool {
	if bucket == "" {
		bucket = model.CheckExecutionBucketMinute
	}

	if len(allowed) == 0 {
		return bucket == model.CheckExecutionBucketMinute ||
			bucket == model.CheckExecutionBucketHour
	}

	for i := range allowed {
		if model.CheckExecutionBucket(allowed[i]) == bucket {
			return true
		}
	}

	return false
}

func validateBucketFilter(filter model.CheckExecutionAggregateFilter) error {
	if filter.ServiceID == "" {
		return fmt.Errorf("service_id is required")
	}

	if filter.CheckID == "" {
		return fmt.Errorf("check_id is required")
	}

	switch filter.Bucket {
	case "", model.CheckExecutionBucketSecond, model.CheckExecutionBucketMinute,
		model.CheckExecutionBucketHour, model.CheckExecutionBucketDay:
	default:
		return fmt.Errorf("unsupported bucket %q", filter.Bucket)
	}

	return nil
}
