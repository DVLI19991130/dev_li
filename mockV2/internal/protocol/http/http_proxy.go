package http

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v3/middleware/proxy"
	"github.com/pkg/errors"
	"mock/internal/slb"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
	"mock/internal/conf"
	"mock/internal/protocol/http/middleware"
)

type ProxyServer struct {
	app *fiber.App
	ctx context.Context

	cfg *conf.Config
}

func (s *ProxyServer) Init() error {
	app := fiber.New(fiber.Config{
		ReduceMemoryUsage: true,
	})

	// Register proxy logging middleware
	app.Use(middleware.ProxyLoggingMiddleware())

	for _, api := range s.cfg.GetHttpProxyAPI() {
		app.Add([]string{api.Method}, api.URL, func(ctx fiber.Ctx) error {
			// Store delay in locals for middleware to read
			ctx.Locals("delay", api.Delay)

			if api.Delay > 0 {
				time.Sleep(time.Duration(api.Delay) * time.Millisecond)
			}

			// Record backend selection info
			backendUrl := slb.WeightedRandom(api.Backend.Addr, api.Backend.Weight)
			ctx.Locals("backend", backendUrl)
			// Weight selection log is recorded in slb.WeightedRandom

			proxyUrl := fmt.Sprintf("http://%s%s", backendUrl, api.URL)
			return proxy.Do(ctx, proxyUrl)
		})
	}

	s.app = app
	return nil
}

func (s *ProxyServer) Start() error {
	go func() {
		<-s.ctx.Done()
		_ = s.Close()

		log.Info().Str("serverName", s.cfg.ServerName).Str("serverMode", s.cfg.ServerMode).
			Str("protocol", s.cfg.Protocol).Int("port", s.cfg.ServerPort).
			Msg("server shutdown")
	}()

	log.Info().Str("serverName", s.cfg.ServerName).Str("serverMode", s.cfg.ServerMode).
		Str("protocol", s.cfg.Protocol).Int("port", s.cfg.ServerPort).
		Msg("server started")
	return s.app.Listen(fmt.Sprintf(":%d", s.cfg.ServerPort))
}

func (s *ProxyServer) Close() error {
	return s.app.Shutdown()
}

// StartProxy starts HTTP Proxy service
func StartProxy(ctx context.Context, cfg *conf.Config) error {
	mockServer := &ProxyServer{
		cfg: cfg,
		ctx: ctx,
	}

	if err := mockServer.Init(); err != nil {
		return errors.Wrap(err, "Failed to initialize HTTP Mock Server")
	}

	return mockServer.Start()
}
