package tcp

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	c "github.com/pixel365/pulse/internal/config"
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
	} else {
		result.Status = model.Success
	}

	return result
}

func (c *Checker) request(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()

	address := net.JoinHostPort(c.config.Spec.Host, fmt.Sprint(c.config.Spec.Port))
	dialer := &net.Dialer{}

	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return fmt.Errorf("could not establish tcp connection: %w", err)
	}

	defer func() {
		_ = conn.Close()
	}()

	if deadline, ok := ctx.Deadline(); ok {
		if err = conn.SetDeadline(deadline); err != nil {
			return fmt.Errorf("could not set tcp deadline: %w", err)
		}
	}

	if c.config.Spec.Send != "" {
		if _, err = io.WriteString(conn, c.config.Spec.Send); err != nil {
			return fmt.Errorf("could not write tcp payload: %w", err)
		}
	}

	if err = checkResponse(conn, c.config.Spec.Expect); err != nil {
		return err
	}

	return nil
}

func checkResponse(conn net.Conn, expect *c.StringExpect) error {
	if expect == nil {
		return nil
	}

	if expect.Equals != "" {
		return checkEquals(conn, expect)
	}

	if expect.Contains != "" {
		return checkContains(conn, expect.Contains)
	}

	return nil
}

func checkEquals(conn net.Conn, expect *c.StringExpect) error {
	body, err := io.ReadAll(conn)
	if err != nil {
		return fmt.Errorf("could not read tcp response: %w", err)
	}

	response := string(body)
	if expect.Contains != "" && !strings.Contains(response, expect.Contains) {
		return fmt.Errorf("tcp response does not contain %q", expect.Contains)
	}

	if response != expect.Equals {
		return fmt.Errorf("tcp response does not equal expected value")
	}

	return nil
}

func checkContains(conn net.Conn, contains string) error {
	var builder strings.Builder
	buf := make([]byte, 4096)

	for {
		n, err := conn.Read(buf)
		if n > 0 {
			builder.Write(buf[:n])
			if strings.Contains(builder.String(), contains) {
				return nil
			}
		}

		if err == nil {
			continue
		}

		if errors.Is(err, io.EOF) {
			break
		}

		return fmt.Errorf("could not read tcp response: %w", err)
	}

	return fmt.Errorf("tcp response does not contain %q", contains)
}
