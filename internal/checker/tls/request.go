package tls

import (
	"context"
	ctls "crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/pixel365/pulse/internal/config"
	"github.com/pixel365/pulse/internal/e"
)

func (c *Checker) request(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()

	spec, err := config.ResolveTLSSpecEnv(c.config.Spec)
	if err != nil {
		return e.NewError(e.ErrInternal, fmt.Sprintf("could not resolve tls spec: %v", err))
	}

	address := net.JoinHostPort(spec.Host, fmt.Sprint(spec.Port))
	serverName := spec.ServerName
	if serverName == "" {
		serverName = spec.Host
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
	if validityLeft < spec.MinValidity {
		return e.NewError(
			e.ErrConstraint,
			fmt.Sprintf(
				"certificate validity %s is below required minimum %s",
				validityLeft.Truncate(time.Second),
				spec.MinValidity,
			),
		)
	}

	return nil
}
