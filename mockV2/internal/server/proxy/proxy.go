package proxy

import (
	"context"
	"github.com/pkg/errors"
	"mock/internal/conf"
	hp "mock/internal/protocol/http"
)

// StartProxyServer starts corresponding Proxy Server based on protocol
func StartProxyServer(ctx context.Context, cfg *conf.Config) error {
	switch cfg.Protocol {
	case "http":
		return hp.StartProxy(ctx, cfg)
	default:
		return errors.Errorf("Unsupported protocol: %s", cfg.Protocol)
	}
}
