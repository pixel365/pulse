package grpc

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"

	"github.com/pixel365/pulse/internal/config"
	"github.com/pixel365/pulse/internal/e"
)

func (c *Checker) request(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()

	spec, err := config.ResolveGRPCSpecEnv(c.config.Spec)
	if err != nil {
		return e.NewError(e.ErrInternal, fmt.Sprintf("could not resolve grpc spec: %v", err))
	}

	address := net.JoinHostPort(spec.Host, fmt.Sprint(spec.Port))
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return e.NewError(
			e.ErrInternal,
			fmt.Sprintf("could not create grpc client: %v", err),
		)
	}

	defer func() {
		_ = conn.Close()
	}()

	if len(spec.Metadata) > 0 {
		ctx = metadata.NewOutgoingContext(ctx, metadata.New(spec.Metadata))
	}

	service := ""
	if spec.Request != nil {
		service = spec.Request.Service
	}

	client := healthpb.NewHealthClient(conn)
	resp, err := client.Check(ctx, &healthpb.HealthCheckRequest{
		Service: service,
	})
	if err != nil {
		return fmt.Errorf("could not execute grpc health check: %w", err)
	}

	if got := resp.GetStatus().String(); got != string(spec.ExpectedHealthStatus) {
		return e.NewError(
			e.ErrConstraint,
			fmt.Sprintf(
				"unexpected grpc health status: expected %s, got %s",
				spec.ExpectedHealthStatus,
				got,
			),
		)
	}

	return nil
}
