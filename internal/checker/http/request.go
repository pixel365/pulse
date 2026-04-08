package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	h "net/http"
	"net/url"
	"strings"

	"github.com/pixel365/pulse/internal/config"
	"github.com/pixel365/pulse/internal/e"
)

func (c *Checker) request(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()

	spec, err := config.ResolveHTTPSpecEnv(c.config.Spec)
	if err != nil {
		return e.NewError(e.ErrInternal, fmt.Sprintf("could not resolve http spec: %v", err))
	}

	cl := &h.Client{
		CheckRedirect: func(req *h.Request, via []*h.Request) error {
			if !spec.FollowRedirects {
				return h.ErrUseLastResponse
			}

			if len(via) >= 10 {
				return fmt.Errorf("stopped after %d redirects", len(via))
			}

			return nil
		},
	}

	req, err := makeRequest(ctx, spec)
	if err != nil {
		return fmt.Errorf("could not make request: %w", err)
	}

	for k, v := range spec.Headers {
		req.Header.Set(k, v)
	}

	res, err := cl.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		_ = res.Body.Close()
	}()

	if err = checkCode(res.StatusCode, spec.SuccessCodes); err != nil {
		return fmt.Errorf("could not check response code: %w", err)
	}

	if err = checkBody(spec.ExpectedBody, res.Body); err != nil {
		return fmt.Errorf("could not parse response body: %w", err)
	}

	return nil
}

func makeRequest(ctx context.Context, spec config.HttpSpec) (*h.Request, error) {
	var req *h.Request

	switch spec.Method {
	case "GET":
		fullUrl := spec.URL
		if len(spec.Payload) > 0 {
			params := url.Values{}
			for k, v := range spec.Payload {
				params.Add(k, fmt.Sprint(v))
			}

			fullUrl = fmt.Sprintf("%s?%s", fullUrl, params.Encode())
		}

		rq, err := h.NewRequestWithContext(ctx, spec.Method, fullUrl, nil)
		if err != nil {
			return nil, fmt.Errorf("could not send request: %w", err)
		}
		req = rq
	case "POST":
		var payload io.Reader
		if len(spec.Payload) > 0 {
			data, err := json.Marshal(spec.Payload)
			if err != nil {
				return nil, fmt.Errorf("could not marshal data: %w", err)
			}
			payload = bytes.NewReader(data)
		}

		rq, err := h.NewRequestWithContext(ctx, spec.Method, spec.URL, payload)
		if err != nil {
			return nil, fmt.Errorf("could not send request: %w", err)
		}
		req = rq
	default:
		return nil, e.NewError(
			e.ErrInternal,
			fmt.Sprintf("unsupported method: %s", spec.Method),
		)
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
		return e.NewError(e.ErrProtocol, fmt.Sprintf("unsuccess code %d", statusCode))
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
			return e.NewError(
				e.ErrConstraint,
				"response body does not contain expected value",
			)
		}

		if equals != "" {
			bodyOk = string(body) == equals
		}
	}

	if !bodyOk {
		return e.NewError(e.ErrConstraint, "response body does not equal expected value")
	}

	return nil
}
