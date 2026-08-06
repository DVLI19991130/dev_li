package nacos

import (
	"fmt"
	"strings"
	"time"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// NacosClient Nacos registry client
type NacosClient struct {
	namingClient interface {
		RegisterInstance(param vo.RegisterInstanceParam) (bool, error)
		DeregisterInstance(param vo.DeregisterInstanceParam) (bool, error)
		SelectAllInstances(param vo.SelectAllInstancesParam) ([]model.Instance, error)
		CloseClient()
	}
}

// NewNacosClient creates Nacos client
// addr: Nacos server address list, e.g. ["127.0.0.1:8848"]
// namespace: namespace, defaults to "public"
// timeout: connection timeout, default 10 seconds
func NewNacosClient(addr []string, namespace string, timeout time.Duration) (*NacosClient, error) {
	if len(addr) == 0 {
		return nil, errors.New("Nacos address cannot be empty")
	}

	if timeout == 0 {
		timeout = 10 * time.Second
	}

	// 构建 ServerConfigs
	serverConfigs := make([]constant.ServerConfig, 0, len(addr))
	for _, a := range addr {
		parts := strings.Split(a, ":")
		ip := parts[0]
		port := uint64(8848)
		if len(parts) > 1 {
			fmt.Sscanf(parts[1], "%d", &port)
		}
		serverConfigs = append(serverConfigs, *constant.NewServerConfig(ip, port))
	}

	// 构建 ClientConfig
	clientConfig := constant.NewClientConfig(
		constant.WithTimeoutMs(uint64(timeout.Milliseconds())),
		constant.WithNamespaceId(namespace),
	)

	// 创建 Naming Client
	namingClient, err := clients.NewNamingClient(vo.NacosClientParam{
		ClientConfig:  clientConfig,
		ServerConfigs: serverConfigs,
	})

	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create Nacos naming client")
	}

	log.Info().
		Strs("addr", addr).
		Str("namespace", namespace).
		Dur("timeout", timeout).
		Msg("[Nacos Registry] Nacos client created successfully")

	return &NacosClient{
		namingClient: namingClient,
	}, nil
}

// RegisterInstance registers service instance to Nacos
func (c *NacosClient) RegisterInstance(serviceName, ip string, port int) error {
	if c.namingClient == nil {
		return errors.New("Nacos client not initialized")
	}

	param := vo.RegisterInstanceParam{
		Ip:          ip,
		Port:        uint64(port),
		ServiceName: serviceName,
		GroupName:   "DEFAULT_GROUP",
		Weight:      1.0,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
	}

	success, err := c.namingClient.RegisterInstance(param)
	if err != nil {
		return errors.Wrapf(err, "Failed to register service instance to Nacos")
	}
	if !success {
		return errors.New("Failed to register service instance to Nacos")
	}

	log.Info().
		Str("serviceName", serviceName).
		Str("ip", ip).
		Int("port", port).
		Msg("[Nacos Registry] Service instance registered successfully")
	return nil
}

// DeregisterInstance deregisters service instance from Nacos
func (c *NacosClient) DeregisterInstance(serviceName, ip string, port int) error {
	if c.namingClient == nil {
		return errors.New("Nacos client not initialized")
	}

	param := vo.DeregisterInstanceParam{
		Ip:          ip,
		Port:        uint64(port),
		ServiceName: serviceName,
		GroupName:   "DEFAULT_GROUP",
		Ephemeral:   true,
	}

	success, err := c.namingClient.DeregisterInstance(param)
	if err != nil {
		return errors.Wrapf(err, "Failed to deregister service instance from Nacos")
	}
	if !success {
		return errors.New("Failed to deregister service instance from Nacos")
	}

	log.Info().
		Str("serviceName", serviceName).
		Str("ip", ip).
		Int("port", port).
		Msg("[Nacos Registry] Service instance deregistered successfully")
	return nil
}

// Close closes Nacos client
func (c *NacosClient) Close() {
	if c.namingClient != nil {
		c.namingClient.CloseClient()
		log.Info().Msg("[Nacos Registry] Nacos client closed")
	}
}
