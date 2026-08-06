package server

import (
	"context"
	"github.com/pkg/errors"
	"mock/internal/conf"
	"mock/internal/server/mock"
	"mock/internal/server/proxy"
)

func StartServer(ctx context.Context, cfg *conf.Config) error {
	switch cfg.ServerMode {
	case "mock":
		return mock.StartMockServer(ctx, cfg)
	case "proxy":
		return proxy.StartProxyServer(ctx, cfg)
	default:
		return errors.Errorf("Unsupported server type: %s", cfg.ServerMode)
	}
}
