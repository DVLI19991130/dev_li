// Package conf - configuration file parsing and validation
package conf

import (
	"encoding/json"
	"fmt"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zhtrans "github.com/go-playground/validator/v10/translations/zh"
	"github.com/pkg/errors"
	"os"
	"strings"
)

// validator instance
var (
	validate *validator.Validate
	trans    ut.Translator
)

func init() {
	// Initialize validator
	z := zh.New()
	uni := ut.New(z, z)

	trans, _ = uni.GetTranslator("zh")
	validate = validator.New()

	_ = zhtrans.RegisterDefaultTranslations(validate, trans)
}

// Load loads and validates configuration from the specified path
func Load(path string) (Configs, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to read config file")
	}

	// Parse JSON
	var configs Configs
	if err := json.Unmarshal(data, &configs); err != nil {
		return nil, errors.Wrapf(err, "Failed to parse JSON")
	}

	// Validate Config
	for _, serverConfig := range configs {
		if err := Validate(serverConfig); err != nil {
			return nil, errors.Wrapf(err, "Failed to validate config")
		}
	}
	return configs, nil
}

// Validate validates configuration completeness
func Validate(c *Config) error {
	// Basic structure validation
	if err := validate.Struct(c); err != nil {
		return printError(err)
	}

	// Validate Registry config
	if err := validateRegistry(c.Registry); err != nil {
		return printError(err)
	}

	// Convert and validate API config based on protocol type
	for i, api := range c.Api {
		converted, err := convertAndValidateAPI(api, c.Protocol, c.ServerMode)
		if err != nil {
			return printError(err)
		}
		c.Api[i] = converted
	}
	return nil
}

// printError formats validation errors
func printError(err error) error {
	var errs validator.ValidationErrors
	ok := errors.As(err, &errs)
	if !ok {
		return err
	}

	var s []string
	for _, e := range errs {
		s = append(s, e.Translate(trans))
	}
	return errors.New(strings.Join(s, ","))
}

// validateRegistry validates registry configuration
func validateRegistry(r *Registry) error {
	if r == nil {
		return nil
	}

	if r.Zookeeper != nil {
		if err := validate.Struct(r.Zookeeper); err != nil {
			return err
		}
	}

	if r.Nacos != nil {
		if err := validate.Struct(r.Nacos); err != nil {
			return err
		}
	}
	return nil
}

// convertAndValidateAPI validates API config and converts type based on protocol
func convertAndValidateAPI(api any, protocol, serverMode string) (any, error) {
	switch protocol {
	case "http":
		if serverMode == "mock" {
			var httpAPI HttpAPI
			data, _ := json.Marshal(api)
			if err := json.Unmarshal(data, &httpAPI); err != nil {
				return nil, errors.Wrap(err, "Failed to parse HTTP API config")
			}

			if err := validate.Struct(httpAPI); err != nil {
				return nil, errors.Wrap(printError(err), "Failed to validate HTTP API config")
			}

			// Pre-compute static response and dynamic flag
			httpAPI.Prepare()
			return &httpAPI, nil
		}

		if serverMode == "proxy" {
			var httpProxyAPI HttpProxyAPI
			data, _ := json.Marshal(api)
			if err := json.Unmarshal(data, &httpProxyAPI); err != nil {
				return nil, errors.Wrap(err, "Failed to parse HTTP Proxy API config")
			}

			if err := validate.Struct(httpProxyAPI); err != nil {
				return nil, errors.Wrap(printError(err), "Failed to validate HTTP Proxy API config")
			}

			if len(httpProxyAPI.Backend.Addr) != len(httpProxyAPI.Backend.Weight) {
				return nil, fmt.Errorf("backend Addr and Weight count mismatch")
			}
			return &httpProxyAPI, nil
		}

		return nil, errors.Errorf("Unknown server type: %s", serverMode)
	case "dubbo":
		if serverMode != "mock" {
			return nil, errors.Errorf("Dubbo protocol only supports mock mode, got: %s", serverMode)
		}

		var dubboAPI DubboAPI
		data, _ := json.Marshal(api)
		if err := json.Unmarshal(data, &dubboAPI); err != nil {
			return nil, errors.Wrap(err, "Failed to parse Dubbo API config")
		}

		if err := validate.Struct(dubboAPI); err != nil {
			return nil, errors.Wrap(printError(err), "Failed to validate Dubbo API config")
		}

		// Validate Dubbo methods
		for _, method := range dubboAPI.Methods {
			if err := validate.Struct(method); err != nil {
				return nil, errors.Wrap(printError(err), "Failed to validate Dubbo API Method config")
			}

			if err := validateResponseTypes(method.ResponseTypes, method.Response, method.Name); err != nil {
				return nil, err
			}

			// Detect dynamic values and build type converters
			method.Prepare()
			method.buildConverter()
		}
		// Build method index for O(1) lookup
		dubboAPI.buildMethodIndex()
		return &dubboAPI, nil
	}

	return nil, errors.Errorf("Unknown protocol type: %s", protocol)
}

// validateResponseTypes validates that responseTypes matches response structure
// responseTypes must exactly match response fields: field count and nested structure
// Note: _class is a special field for specifying Java class name, not required in response
func validateResponseTypes(types, resp map[string]any, methodName string) error {
	// Check that fields in responseTypes exist in response (skip _class)
	for key := range types {
		if key == "_class" {
			continue
		}
		if _, ok := resp[key]; !ok {
			return fmt.Errorf("Dubbo method [%s]: field [%s] in responseTypes does not exist in response", methodName, key)
		}
	}

	// Check that fields in response are declared in responseTypes
	for key := range resp {
		if _, ok := types[key]; !ok {
			return fmt.Errorf("Dubbo method [%s]: field [%s] in response is not declared in responseTypes", methodName, key)
		}
	}

	// Recursively validate nested objects
	for key, typeVal := range types {
		if key == "_class" {
			continue
		}
		respVal := resp[key]
		if typeMap, ok := typeVal.(map[string]any); ok {
			if respMap, ok := respVal.(map[string]any); ok {
				if err := validateResponseTypes(typeMap, respMap, methodName+"."+key); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
