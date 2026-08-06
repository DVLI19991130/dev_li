// init command - initialize mock.json configuration file
package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"mock/internal/conf"

	"github.com/spf13/cobra"
)

var (
	serverMode string
	protocol   string
	registry   string
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize mock.json configuration file",
	Long:  `Generate default mock.json configuration file in current directory or specified path.`,
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVarP(&serverMode, "serverMode", "s", "", "Server mode (mock|proxy)")
	initCmd.Flags().StringVarP(&protocol, "protocol", "p", "", "Protocol type (http|dubbo)")
	initCmd.Flags().StringVarP(&registry, "registry", "r", "", "Registry type (nacos|zookeeper)")
}

func runInit(cmd *cobra.Command, args []string) error {
	if serverMode == "" {
		return fmt.Errorf("required flag(s) serverMode not set")
	}

	if protocol == "" {
		return fmt.Errorf("required flag(s) protocol not set")
	}

	config, err := generateConfig()
	if err != nil {
		return err
	}

	configs, err := readOrCreateConfigs()
	if err != nil {
		return fmt.Errorf("failed to read existing config: %w", err)
	}

	for _, c := range configs {
		if c.ServerMode == serverMode && c.Protocol == protocol && getRegistryType(c) == registry {
			return fmt.Errorf("configuration for serverMode=%s, protocol=%s, registry=%s already exists", serverMode, protocol, registry)
		}
	}

	configs = append(configs, config)

	if err := writeConfigs(configs); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	fmt.Printf("Successfully generated mock.json with %s + %s configuration\n", serverMode, protocol)
	return nil
}

func getRegistryType(c *conf.Config) string {
	if c.Registry == nil {
		return ""
	}

	if c.Registry.Nacos != nil {
		return "nacos"
	}

	if c.Registry.Zookeeper != nil {
		return "zookeeper"
	}
	return ""
}

func generateConfig() (*conf.Config, error) {
	config := &conf.Config{
		ServerMode: serverMode,
		Protocol:   protocol,
	}

	switch serverMode + "/" + protocol {
	case "mock/http":
		config.ServerName = "mock-http-server"
		config.ServerPort = 8080

		if registry != "" {
			if registry != "nacos" {
				return nil, fmt.Errorf("zookeeper registry is only supported for dubbo protocol")
			}

			config.Registry = &conf.Registry{
				Nacos: &conf.Nacos{
					Addr:      []string{"127.0.0.1:8848"},
					Namespace: "public",
				},
			}
		}

		config.Api = []any{
			&conf.HttpAPI{
				URL:    "/example/api",
				Method: "GET",
				Delay:  -1,
				Response: map[string]any{
					"resCode":    "0000",
					"resMessage": "success",
				},
			},
		}

	case "mock/dubbo":
		config.ServerName = "mock-dubbo-server"
		config.ServerPort = 8088

		config.Registry = &conf.Registry{
			Zookeeper: &conf.Zookeeper{
				Addr: []string{"127.0.0.1:2181"},
			},
		}

		config.Api = []any{
			&conf.DubboAPI{
				Interface:     "com.example.service.UserService",
				Version:       "1.0.0",
				Serialization: "hessian2",
				Registry:      registry,
				Methods: []*conf.DubboMethod{
					{
						Name:      "getUser",
						Delay:     -1,
						ParamType: []string{"java.lang.String"},
						ResponseTypes: map[string]any{
							"userId":   "string",
							"userName": "string",
						},
						Response: map[string]any{
							"userId":   "1001",
							"userName": "test",
						},
					},
				},
			},
		}

	case "proxy/http":
		config.ServerName = "proxy-http-server"
		config.ServerPort = 8080
		config.Api = []any{
			&conf.HttpProxyAPI{
				URL:    "/example/api",
				Method: "GET",
				Backend: &conf.Backend{
					Addr:   []string{"127.0.0.1:8080"},
					Weight: []int{100},
				},
				Delay: -1,
			},
		}
	default:
		return nil, fmt.Errorf("unsupported serverMode=%s, protocol=%s combination", serverMode, protocol)
	}

	return config, nil
}

func readOrCreateConfigs() (conf.Configs, error) {
	data, err := os.ReadFile("mock.json")
	if err != nil {
		if os.IsNotExist(err) {
			return conf.Configs{}, nil
		}
		return nil, err
	}

	var configs conf.Configs
	if err := json.Unmarshal(data, &configs); err != nil {
		return nil, fmt.Errorf("failed to parse existing mock.json: %w", err)
	}

	return configs, nil
}

func writeConfigs(configs conf.Configs) error {
	data, err := json.MarshalIndent(configs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile("mock.json", data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
