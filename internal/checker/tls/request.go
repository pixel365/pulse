package tls

import (
	"context"
	ctls "crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/pixel365/pulse/internal/e"
)

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
		return e.NewError(
			e.ErrInternal,
			fmt.Sprintf("unexpected connection type: %T", conn),
		)
	}

	defer func() {
		_ = tlsConn.Close()
	}()

	state := tlsConn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return e.NewError(e.ErrProtocol, "no peer certificates presented")
	}

	leaf := state.PeerCertificates[0]
	validityLeft := time.Until(leaf.NotAfter)
	if validityLeft < c.config.Spec.MinValidity {
		return e.NewError(
			e.ErrConstraint,
			fmt.Sprintf(
				"certificate validity %s is below required minimum %s",
				validityLeft.Truncate(time.Second),
				c.config.Spec.MinValidity,
			),
		)
	}

	return nil
}
