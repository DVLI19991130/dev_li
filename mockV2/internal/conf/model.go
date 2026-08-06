package conf

import (
	"encoding/json"
	"mock/internal/dynamic"
	"strings"
)

type Configs []*Config

// Config defines the configuration structure
type Config struct {
	ServerName string    `json:"serverName" validate:"required"`
	ServerPort int       `json:"serverPort" validate:"required,min=1,max=65535"`
	ServerMode string    `json:"serverMode" validate:"required,oneof=mock proxy"`
	Protocol   string    `json:"protocol" validate:"required,oneof=http dubbo"`
	Registry   *Registry `json:"registry,omitempty"`
	Api        []any     `json:"api" validate:"required,min=1"`
}

func (c *Config) String() string {
	bytes, _ := json.Marshal(c)
	return string(bytes)
}

func (c *Config) GetHttpAPI() []*HttpAPI {
	result := make([]*HttpAPI, 0, len(c.Api))
	for _, api := range c.Api {
		if v, ok := api.(*HttpAPI); ok {
			result = append(result, v)
		}
	}
	return result
}

func (c *Config) GetHttpProxyAPI() []*HttpProxyAPI {
	result := make([]*HttpProxyAPI, 0, len(c.Api))
	for _, api := range c.Api {
		if v, ok := api.(*HttpProxyAPI); ok {
			result = append(result, v)
		}
	}
	return result
}

func (c *Config) GetDubboAPI() []*DubboAPI {
	result := make([]*DubboAPI, 0, len(c.Api))
	for _, api := range c.Api {
		if v, ok := api.(*DubboAPI); ok {
			result = append(result, v)
		}
	}
	return result
}

// HttpAPI HTTP Mock API configuration
type HttpAPI struct {
	URL      string `json:"url" validate:"required"`
	Method   string `json:"method" validate:"required,oneof=GET POST PUT DELETE PATCH HEAD OPTIONS"`
	Delay    int    `json:"delay" validate:"gte=-1"`
	Response any    `json:"response" validate:"required"`

	// Pre-computed fields to avoid repeated checks during requests
	isDynamic  bool
	staticJSON []byte
}

func (h *HttpAPI) IsDynamic() bool {
	return h.isDynamic
}

func (h *HttpAPI) StaticJSON() []byte {
	return h.staticJSON
}

// Prepare pre-computes static response and dynamic flag
func (h *HttpAPI) Prepare() {
	if dynamic.HasDynamicValue(h.Response) {
		h.isDynamic = true
	} else {
		h.isDynamic = false
		h.staticJSON, _ = json.Marshal(h.Response)
	}
}

// HttpProxyAPI HTTP Proxy API configuration
type HttpProxyAPI struct {
	URL     string   `json:"url" validate:"required"`
	Method  string   `json:"method" validate:"required,oneof=GET POST PUT DELETE PATCH HEAD OPTIONS"`
	Backend *Backend `json:"backend" validate:"required"`
	Delay   int      `json:"delay" validate:"gte=-1"`
}

// Backend service configuration
type Backend struct {
	Addr   []string `json:"addr" validate:"required,min=1"`
	Weight []int    `json:"weight" validate:"omitempty,min=1,dive,min=0,max=100"`
}

// DubboAPI Dubbo API configuration
type DubboAPI struct {
	Interface     string         `json:"interface" validate:"required"`
	Version       string         `json:"version" validate:"required"`
	Group         string         `json:"group"`
	Serialization string         `json:"serialization" validate:"required,oneof=hessian2"`
	Delay         int            `json:"delay" validate:"gte=-1"`
	Registry      string         `json:"registry" validate:"omitempty,oneof=zookeeper"`
	Methods       []*DubboMethod `json:"methods"`

	methodIndex map[string]*DubboMethod // key: methodName|paramType1,paramType2,...
}

func (d *DubboAPI) FindMethod(method string) *DubboMethod {
	return d.methodIndex[method]
}

// DubboMethod Dubbo method configuration
type DubboMethod struct {
	Name          string         `json:"name" validate:"required"`
	Delay         int            `json:"delay" validate:"gte=-1"`
	Timeout       int            `json:"timeout" validate:"gte=0"`
	ParamType     []string       `json:"paramType" validate:"required,min=1"`
	Response      map[string]any `json:"response" validate:"required"`
	ResponseTypes map[string]any `json:"responseTypes" validate:"required"` // Java type declarations for Hessian2 serialization control

	// Pre-built type converter functions
	converter map[string]func(any) any
	// Cache fields (detection only, no StaticJSON)
	isDynamic bool
}

func (h *DubboMethod) Converter() map[string]func(any) any {
	return h.converter
}

func (h *DubboMethod) IsDynamic() bool {
	return h.isDynamic
}

// Prepare pre-computes static response and dynamic flag
func (h *DubboMethod) Prepare() {
	if dynamic.HasDynamicValue(h.Response) {
		h.isDynamic = true
	} else {
		h.isDynamic = false
	}
}

// buildConverter pre-builds type converter functions from ResponseTypes
func (h *DubboMethod) buildConverter() {
	h.converter = buildResponseConverter(h.ResponseTypes)
}

// buildMethodIndex builds method index map for O(1) lookup
func (d *DubboAPI) buildMethodIndex() {
	d.methodIndex = make(map[string]*DubboMethod, len(d.Methods))
	for _, m := range d.Methods {
		key := methodKey(m.Name, m.ParamType)
		d.methodIndex[key] = m
	}
}

// methodKey builds method index key: methodName|paramType1,paramType2,...
func methodKey(methodName string, paramTypes []string) string {
	return methodName + "|" + strings.Join(paramTypes, ",")
}

// buildResponseConverter pre-builds type converter function map
func buildResponseConverter(types map[string]any) map[string]func(any) any {
	if types == nil {
		return nil
	}
	converters := make(map[string]func(any) any)
	for key, typeVal := range types {
		if key == "_class" {
			continue
		}
		converters[key] = buildConverter(typeVal)
	}
	return converters
}

// buildConverter builds converter function for a single type
func buildConverter(typeVal any) func(any) any {
	// If it's a nested object structure, the converter function needs to recursively call applyResponseTypes
	if typeMap, ok := typeVal.(map[string]any); ok {
		return func(val any) any {
			if valMap, ok := val.(map[string]any); ok {
				result := applyResponseTypes(valMap, typeMap)
				// If _class is in the type declaration, add it to the result for Hessian serialization
				if className, ok := typeMap["_class"].(string); ok {
					result["_class"] = className
				}
				return result
			}
			return val
		}
	}

	// Basic types, use pre-computed conversion
	t, ok := typeVal.(string)
	if !ok {
		return func(v any) any { return v }
	}

	switch t {
	case "int":
		return func(v any) any {
			if f, ok := v.(float64); ok {
				return int32(f)
			}
			return v
		}
	case "long":
		return func(v any) any {
			if f, ok := v.(float64); ok {
				return int64(f)
			}
			return v
		}
	case "double", "string", "bool":
		return func(v any) any { return v }
	default:
		return func(v any) any { return v }
	}
}

// applyResponseTypes recursively converts response types according to responseTypes
// Converts JSON-parsed float64 to types declared in responseTypes
func applyResponseTypes(resp, types map[string]any) map[string]any {
	result := make(map[string]any)
	for key, val := range resp {
		typeVal, ok := types[key]
		if !ok {
			result[key] = val
			continue
		}
		result[key] = convertByType(val, typeVal)
	}
	return result
}

// convertByType converts value according to type declaration
func convertByType(val, typeVal any) any {
	// If type declaration is a nested object structure
	if typeMap, ok := typeVal.(map[string]any); ok {
		if valMap, ok := val.(map[string]any); ok {
			result := applyResponseTypes(valMap, typeMap)
			// If _class is in the type declaration, add it to the result for Hessian serialization
			if className, ok := typeMap["_class"].(string); ok {
				result["_class"] = className
			}
			return result
		}
	}

	t, ok := typeVal.(string)
	if !ok {
		return val
	}

	switch t {
	case "int":
		if f, ok := val.(float64); ok {
			return int32(f)
		}
	case "long":
		if f, ok := val.(float64); ok {
			return int64(f)
		}
	case "double":
		// Already float64
	case "string":
		// Already string
	case "bool":
		// Already bool
	}
	return val
}

// Registry registry configuration
type Registry struct {
	Zookeeper *Zookeeper `json:"zookeeper,omitempty"`
	Nacos     *Nacos     `json:"nacos,omitempty"`
}

// Zookeeper Zookeeper configuration
type Zookeeper struct {
	Addr []string `json:"addr" validate:"required,min=1"`
}

// Nacos Nacos configuration
type Nacos struct {
	Addr      []string `json:"addr" validate:"required,min=1"`
	Namespace string   `json:"namespace"`
}
