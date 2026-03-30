package http

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	h "net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pixel365/pulse/internal/config"
	"github.com/pixel365/pulse/internal/model"
)

func (c *Checker) execute(ctx context.Context) model.CheckExecutionResult {
	var err error

	result := model.CheckExecutionResult{
		ExecutionID: rand.Text(),
		CheckID:     c.config.ID,
		ServiceID:   c.config.Service,
		CheckType:   c.config.Type,
		Status:      model.Success,
		StartedAt:   time.Now().UTC(),
	}

	attempts := c.config.Retries + 1
	for i := 0; i < attempts; i++ {
		result.AttemptsTotal = i + 1

		if ctx.Err() != nil {
			err = ctx.Err()
			break
		}

		err = nil
		reqErr := c.request(ctx)
		if reqErr != nil {
			err = reqErr
			continue
		}
		break
	}

	result.FinishedAt = time.Now().UTC()
	result.Duration = result.FinishedAt.Sub(result.StartedAt)

	if err != nil {
		result.Status = model.Failure
		switch {
		case errors.Is(err, context.Canceled),
			errors.Is(err, context.DeadlineExceeded),
			errors.Is(err, ErrCtxCancelled),
			errors.Is(err, ErrTimeout):
			result.ErrorKind = model.ErrTimeout
		case errors.Is(err, ErrResponseBody):
			result.ErrorKind = model.ErrResponseBody
		case errors.Is(err, ErrCode):
			result.ErrorKind = model.ErrStatusCode
		default:
			result.ErrorKind = model.ErrUnknown
		}
	} else {
		result.Status = model.Success
	}

	return result
}

func (c *Checker) request(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()

	cl := &h.Client{
		CheckRedirect: func(req *h.Request, via []*h.Request) error {
			if !c.config.Spec.FollowRedirects {
				return h.ErrUseLastResponse
			}

			if len(via) >= 10 {
				return fmt.Errorf("stopped after %d redirects", len(via))
			}

			return nil
		},
	}

	req, err := makeRequest(ctx, c.config)
	if err != nil {
		return fmt.Errorf("could not make request: %w", err)
	}

	for k, v := range c.config.Spec.Headers {
		req.Header.Set(k, v)
	}

	res, err := cl.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		_ = res.Body.Close()
	}()

	if err = checkCode(res.StatusCode, c.config.Spec.SuccessCodes); err != nil {
		return fmt.Errorf("could not check response code: %w", err)
	}

	if err = checkBody(c.config.Spec.ExpectedBody, res.Body); err != nil {
		return fmt.Errorf("could not parse response body: %w", err)
	}

	return nil
}

func makeRequest(ctx context.Context, config Alias) (*h.Request, error) {
	var req *h.Request

	switch config.Spec.Method {
	case "GET":
		fullUrl := config.Spec.URL
		if len(config.Spec.Payload) > 0 {
			params := url.Values{}
			for k, v := range config.Spec.Payload {
				params.Add(k, fmt.Sprint(v))
			}

			fullUrl = fmt.Sprintf("%s?%s", fullUrl, params.Encode())
		}

		rq, err := h.NewRequestWithContext(ctx, config.Spec.Method, fullUrl, nil)
		if err != nil {
			return nil, fmt.Errorf("could not send request: %w", err)
		}
		req = rq
	case "POST":
		var payload io.Reader
		if len(config.Spec.Payload) > 0 {
			data, err := json.Marshal(config.Spec.Payload)
			if err != nil {
				return nil, fmt.Errorf("could not marshal data: %w", err)
			}
			payload = bytes.NewReader(data)
		}

		rq, err := h.NewRequestWithContext(ctx, config.Spec.Method, config.Spec.URL, payload)
		if err != nil {
			return nil, fmt.Errorf("could not send request: %w", err)
		}
		req = rq
	default:
		return nil, fmt.Errorf("unsupported method: %s", config.Spec.Method)
	}

	return req, nil
}

func checkCode(statusCode int, codes []int) error {
	var success bool
	for _, code := range codes {
		if statusCode == code {
			success = true
			break
		}
	}

	if !success {
		return fmt.Errorf("%w %d", ErrCode, statusCode)
	}

	return nil
}

func checkBody(expect *config.StringExpect, res io.ReadCloser) error {
	bodyOk := true
	if expect != nil {
		contains := expect.Contains
		equals := expect.Equals

		body, err := io.ReadAll(res)
		if err != nil {
			return fmt.Errorf("could not read response: %w", err)
		}

		if contains != "" {
			bodyOk = strings.Contains(string(body), contains)
		}

		if !bodyOk {
			return ErrResponseBody
		}

		if equals != "" {
			bodyOk = string(body) == equals
		}
	}

	if !bodyOk {
		return ErrResponseBody
	}

	return nil
}
