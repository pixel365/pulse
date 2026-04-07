package api

import (
	"fmt"
	"net/http"

	"github.com/pixel365/pulse/internal/config"
)

func (h *Handler) ensureService(serviceID string) error {
	if _, ok := h.cfgProvider.Current().Services[serviceID]; ok {
		return nil
	}

	return apiError{
		status: http.StatusNotFound,
		err:    fmt.Errorf("service %s not found", serviceID),
	}
}

func (h *Handler) ensureCheck(serviceID, checkID string) error {
	_, err := h.checkFields(serviceID, checkID)
	return err
}

func (h *Handler) checkFields(serviceID, checkID string) (config.CheckFields, error) {
	cfg := h.cfgProvider.Current()

	if fields, ok := checkFields(cfg.HttpChecks, serviceID, checkID); ok {
		return fields, nil
	}

	if fields, ok := checkFields(cfg.TCPChecks, serviceID, checkID); ok {
		return fields, nil
	}

	if fields, ok := checkFields(cfg.GRPCChecks, serviceID, checkID); ok {
		return fields, nil
	}

	if fields, ok := checkFields(cfg.DNSChecks, serviceID, checkID); ok {
		return fields, nil
	}

	if fields, ok := checkFields(cfg.TLSChecks, serviceID, checkID); ok {
		return fields, nil
	}

	return config.CheckFields{}, apiError{
		status: http.StatusNotFound,
		err:    fmt.Errorf("check %s/%s not found", serviceID, checkID),
	}
}

func checkFields[T any](
	checks map[string]config.TypedCheck[T],
	serviceID string,
	checkID string,
) (config.CheckFields, bool) {
	for i := range checks {
		if checks[i].Service == serviceID && checks[i].ID == checkID {
			return checks[i].CheckFields, true
		}
	}

	return config.CheckFields{}, false
}
