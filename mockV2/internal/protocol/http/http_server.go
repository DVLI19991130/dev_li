package http

import (
	"context"
	"fmt"
	"mock/pkg"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"mock/internal/conf"
	"mock/internal/dynamic"
	"mock/internal/protocol/http/middleware"
	"mock/internal/registry/nacos"
)

type MockServer struct {
	app *fiber.App

	ctx context.Context
	cfg *conf.Config

	registryClient *nacos.NacosClient
}

func (s *MockServer) Init() error {
	// init http server
	if err := s.initServer(); err != nil {
		return errors.Wrap(err, "http server init fail")
	}

	// init registry
	if err := s.initRegistry(); err != nil {
		return errors.Wrap(err, "registry init failed")
	}
	return nil
}

func (s *MockServer) initServer() error {
	// init fiber app
	app := fiber.New(fiber.Config{
		ReduceMemoryUsage: true,
	})

	// use middleware
	app.Use(middleware.RequestLoggingMiddleware())

	// add http handler
	for _, api := range s.cfg.GetHttpAPI() {
		app.Add([]string{api.Method}, api.URL, func(ctx fiber.Ctx) error {
			// Store delay in locals for middleware to read
			ctx.Locals("delay", api.Delay)

			if api.Delay > 0 {
				time.Sleep(time.Duration(api.Delay) * time.Millisecond)
			}

			if !api.IsDynamic() {
				// Static response: return pre-serialized cache directly
				ctx.Set("Content-Type", "application/json")
				ctx.Response().SetBody(api.StaticJSON())
				return nil
			}

			// Dynamic response: deep copy + process dynamic values
			resData := dynamic.ProcessValue(api.Response, true)
			return ctx.JSON(resData)
		})
	}

	s.app = app
	return nil
}

func (s *MockServer) initRegistry() error {
	if s.cfg.Registry == nil || s.cfg.Registry.Nacos == nil {
		return nil
	}

	registryClient, err := nacos.NewNacosClient(
		s.cfg.Registry.Nacos.Addr,
		s.cfg.Registry.Nacos.Namespace,
		10*time.Second,
	)

	if err != nil {
		return errors.Wrap(err, "Failed to create Nacos client")
	}

	s.registryClient = registryClient
	return nil
}

func (s *MockServer) Start() error {
	if err := s.registerService(); err != nil {
		return errors.Wrap(err, "register server failed")
	}

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

func (s *MockServer) Close() error {
	if s.registryClient != nil {
		s.unregisterAll()
		s.registryClient.Close()
	}

	return s.app.Shutdown()
}

// StartServer starts HTTP Mock service
func StartServer(ctx context.Context, cfg *conf.Config) error {
	mockServer := &MockServer{
		ctx: ctx,
		cfg: cfg,
	}

	if err := mockServer.Init(); err != nil {
		return errors.Wrap(err, "Failed to initialize HTTP Mock Server")
	}

	return mockServer.Start()
}

func (s *MockServer) registerService() error {
	if s.registryClient == nil {
		return nil
	}

	localIP := pkg.GetLocalIP()
	if err := s.registryClient.RegisterInstance(s.cfg.ServerName, localIP, s.cfg.ServerPort); err != nil {
		return errors.Wrap(err, "register nacos failed")
	}

	log.Info().Msg("[HTTP Mock] Service registered to Nacos")
	return nil
}

func (s *MockServer) unregisterAll() {
	if s.registryClient != nil {
		localIP := pkg.GetLocalIP()
		if err := s.registryClient.DeregisterInstance(s.cfg.ServerName, localIP, s.cfg.ServerPort); err != nil {
			log.Error().Err(err).Msg("Failed to deregister service instance from Nacos")
		}
	}
}
