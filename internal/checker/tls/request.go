package tls

import (
	"context"
	"crypto/rand"
	ctls "crypto/tls"
	"fmt"
	"net"
	"time"

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
	serverName := c.config.Spec.ServerName
	if serverName == "" {
		serverName = c.config.Spec.Host
	}

	dialer := &ctls.Dialer{
		Config: &ctls.Config{
			ServerName: serverName,
			MinVersion: ctls.VersionTLS12,
		},
	}

	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return fmt.Errorf("could not establish tls connection: %w", err)
	}

	tlsConn, ok := conn.(*ctls.Conn)
	if !ok {
		_ = conn.Close()
		return fmt.Errorf("unexpected connection type: %T", conn)
	}

	defer func() {
		_ = tlsConn.Close()
	}()

	state := tlsConn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return fmt.Errorf("no peer certificates presented")
	}

	leaf := state.PeerCertificates[0]
	validityLeft := time.Until(leaf.NotAfter)
	if validityLeft < c.config.Spec.MinValidity {
		return fmt.Errorf(
			"certificate validity %s is below required minimum %s",
			validityLeft.Truncate(time.Second),
			c.config.Spec.MinValidity,
		)
	}

	return nil
}
