package mock

import (
	"context"
	"github.com/pkg/errors"
	"mock/internal/conf"
	dp "mock/internal/protocol/dubbo"
	hb "mock/internal/protocol/http"
)

// StartMockServer starts corresponding Mock Server based on protocol
func StartMockServer(ctx context.Context, cfg *conf.Config) error {
	switch cfg.Protocol {
	case "http":
		return hb.StartServer(ctx, cfg)
	case "dubbo":
		return dp.StartServer(ctx, cfg)
	default:
		return errors.Errorf("Unsupported protocol: %s", cfg.Protocol)
	}
}
