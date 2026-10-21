package app

import (
	"context"

	"auth/internal/transport"

	"golang.org/x/sync/errgroup"
)

func runLifecycle(ctx context.Context, c *Container) error {
	c.Logger.Info("application lifecycle starting")

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		c.Logger.Info("starting gRPC server", "addr", ":50053")

		err := transport.StartGRPCServer(ctx, ":50053", c.authService)

		if err != nil {
			c.Logger.Error("gRPC server exited with error", "error", err)
		} else {
			c.Logger.Info("gRPC server stopped")
		}

		return err
	})

	g.Go(func() error {
		c.Logger.Info("starting HTTP server", "addr", ":8080")

		err := transport.StartHTTPServer(ctx, ":8080")

		if err != nil {
			c.Logger.Error("HTTP server exited with error", "error", err)
		} else {
			c.Logger.Info("HTTP server stopped")
		}

		return err
	})

	err := g.Wait()

	if err != nil {
		c.Logger.Error("application lifecycle stopped with error", "error", err)
	} else {
		c.Logger.Info("application lifecycle stopped cleanly")
	}

	return err
}
