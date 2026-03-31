package tcp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/pixel365/pulse/internal/config"
)

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

func checkResponse(conn net.Conn, expect *config.StringExpect) error {
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

func checkEquals(conn net.Conn, expect *config.StringExpect) error {
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
