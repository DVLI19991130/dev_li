package dubbo

import (
	"context"
	"fmt"
	"mock/internal/protocol/dubbo/middleware"
	"mock/pkg"
	"strings"
	"time"

	"dubbo.apache.org/dubbo-go/v3/common"
	"dubbo.apache.org/dubbo-go/v3/common/constant"
	"dubbo.apache.org/dubbo-go/v3/common/extension"
	"dubbo.apache.org/dubbo-go/v3/protocol"
	"dubbo.apache.org/dubbo-go/v3/protocol/invocation"
	"dubbo.apache.org/dubbo-go/v3/registry"
	_ "dubbo.apache.org/dubbo-go/v3/registry/zookeeper"
	"dubbo.apache.org/dubbo-go/v3/remoting/getty"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"mock/internal/conf"
	"mock/internal/dynamic"
)

// MockServer Dubbo Mock service
type MockServer struct {
	ctx context.Context
	cfg *conf.Config

	services map[string]*conf.DubboAPI // key: interfaceName

	registry registry.Registry
	server   *getty.Server

	handler func(*invocation.RPCInvocation) protocol.RPCResult

	done chan struct{}
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

func (s *MockServer) Use(middle func(middleware.DubboHandler) middleware.DubboHandler) {
	s.handler = middle(s.handler)
}

// StartServer starts Dubbo Mock service
func StartServer(ctx context.Context, cfg *conf.Config) error {
	// Create MockServer
	mockServer := &MockServer{
		ctx:      ctx,
		cfg:      cfg,
		services: make(map[string]*conf.DubboAPI),
		done:     make(chan struct{}),
	}

	// load services
	mockServer.loadServices()
	// Request handler
	mockServer.handler = mockHandler(mockServer)
	// Middleware
	mockServer.Use(middleware.RequestLoggingMiddleware())

	// init server
	if err := mockServer.Init(); err != nil {
		return errors.Wrap(err, "Failed to initialize Dubbo Mock Server")
	}

	return mockServer.Start()
}

func (s *MockServer) Start() error {
	// register service
	if err := s.registerService(); err != nil {
		log.Error().Err(err).Msg("注册服务失败")
	}

	go func() {
		<-s.ctx.Done()
		s.Close()

		log.Info().Str("serverName", s.cfg.ServerName).Str("serverMode", s.cfg.ServerMode).
			Str("protocol", s.cfg.Protocol).Int("port", s.cfg.ServerPort).
			Msg("server shutdown")
	}()

	// start getty server
	s.server.Start()

	log.Info().Str("serverName", s.cfg.ServerName).Str("serverMode", s.cfg.ServerMode).
		Str("protocol", s.cfg.Protocol).Int("port", s.cfg.ServerPort).
		Msg("server started")
	<-s.done
	return nil
}

func (s *MockServer) Close() {
	// Close getty.Server
	if s.server != nil {
		s.server.Stop()
		log.Info().Msgf("Getty server closed")
	}

	// Close registry
	if s.registry != nil {
		s.registry.Destroy()
		log.Info().Msgf("Registry connection closed")
	}

	close(s.done)
}

func (s *MockServer) initServer() error {
	localIP := pkg.GetLocalIP()
	addr := fmt.Sprintf("%s:%d", localIP, s.cfg.ServerPort)

	// build dubbo-go URL
	url, err := common.NewURL(
		fmt.Sprintf("dubbo://%s:%d/", localIP, s.cfg.ServerPort),
		common.WithProtocol("dubbo"),
		common.WithIp(localIP),
		common.WithPort(fmt.Sprintf("%d", s.cfg.ServerPort)),
	)

	if err != nil {
		return fmt.Errorf("Failed to build dubbo-go url: %w", err)
	}

	url.Location = addr

	s.server = getty.NewServer(url, s.handler)
	return nil
}

func (s *MockServer) initRegistry() error {
	if s.cfg.Registry == nil || s.cfg.Registry.Zookeeper == nil {
		return nil
	}

	zkConfig := s.cfg.Registry.Zookeeper
	zkAddr := strings.Join(zkConfig.Addr, ",")

	// build registry URL
	registryURL, err := common.NewURL(
		"zookeeper://"+zkAddr+"/?registry.group=dubbo",
		common.WithProtocol("zookeeper"),
		common.WithLocation(zkAddr),
		common.WithPath("/"),
	)

	if err != nil {
		return errors.Wrap(err, "Failed to build zookeeper registry URL")
	}

	// Get native zookeeper registry instance
	s.registry, err = extension.GetRegistry("zookeeper", registryURL)
	if err != nil {
		return errors.Wrap(err, "Failed to get zookeeper registry")
	}
	return nil
}

func (s *MockServer) loadServices() {
	for _, svc := range s.cfg.GetDubboAPI() {
		s.services[svc.Interface] = svc
	}
}

// mockHandler creates Mock request handler (core business logic, without logging)
func mockHandler(server *MockServer) func(*invocation.RPCInvocation) protocol.RPCResult {
	return func(in *invocation.RPCInvocation) protocol.RPCResult {
		result := protocol.RPCResult{}

		// Get interface/method/paramType
		path := in.Attachments()[constant.PathKey].(string)
		methodName := in.MethodName()
		reqParamTypes := in.ParameterTypeNames()

		svc := server.services[path]
		if svc == nil {
			result.SetError(fmt.Errorf("service not found: %s", path))
			return result
		}

		// Method match: find by index O(1)
		method := svc.FindMethod(methodKey(methodName, reqParamTypes))
		if method == nil {
			result.SetError(fmt.Errorf("method not found: %s", methodName))
			return result
		}

		// Handle delay
		if method.Delay > 0 {
			time.Sleep(time.Duration(method.Delay) * time.Millisecond)
		}

		// 1. Type conversion using pre-built converter
		converted := convertWithPrebuilt(method.Response, method.Converter())

		// 2. Dynamic value processing
		if method.IsDynamic() {
			converted = dynamic.ProcessValue(converted, true).(map[string]any)
		}

		result.SetResult(converted)
		return result
	}
}
func (s *MockServer) registerService() error {
	// Register all services to registry
	for _, svc := range s.services {
		methods := make([]string, 0, len(svc.Methods))
		for _, m := range svc.Methods {
			methods = append(methods, m.Name)
		}

		// Build provider URL
		providerURL, err := common.NewURL(
			fmt.Sprintf("dubbo://%s:%d/%s?interface=%s&version=%s&group=%s&serialization=%s&methods=%s",
				pkg.GetLocalIP(), s.cfg.ServerPort, svc.Interface,
				svc.Interface, svc.Version, svc.Group, svc.Serialization,
				strings.Join(methods, ",")),
			common.WithProtocol("dubbo"),
			common.WithIp(pkg.GetLocalIP()),
			common.WithPort(fmt.Sprintf("%d", s.cfg.ServerPort)),
			common.WithPath(svc.Interface),
			common.WithParamsValue(constant.RegistryRoleKey, fmt.Sprintf("%d", common.PROVIDER)),
			common.WithMethods(methods),
		)
		if err != nil {
			return fmt.Errorf("Failed to build provider URL: %w", err)
		}

		// Register service
		if err := s.registry.Register(providerURL); err != nil {
			return fmt.Errorf("Failed to register service: %w", err)
		}

		log.Info().Msgf("[Dubbo Mock] Service registered to ZK: %s", providerURL.String())
	}
	return nil
}

func (s *MockServer) unregisterAll() {
	if s.registry != nil {
		s.registry.Destroy()
		log.Info().Msgf("[Dubbo Mock] Registry connection closed")
	}
}

// methodKey builds method index key: methodName|paramType1,paramType2,...
func methodKey(methodName string, paramTypes []string) string {
	return methodName + "|" + strings.Join(paramTypes, ",")
}

// convertWithPrebuilt converts response using pre-built converter functions
func convertWithPrebuilt(resp map[string]any, converters map[string]func(any) any) map[string]any {
	result := make(map[string]any, len(resp))
	for key, val := range resp {
		if conv, ok := converters[key]; ok {
			result[key] = conv(val)
		} else {
			result[key] = val
		}
	}
	return result
}
