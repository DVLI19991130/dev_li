// Package dynamic - dynamic value generator
// Supports $(funcName,arg1,arg2) syntax to call registered functions for generating dynamic values
package dynamic

import (
	"mock/internal/dynamic/funcs"
	"regexp"
	"strings"
)

// GeneratorHandler dynamic generator function type
type GeneratorHandler func(args ...string) string

// generatorHandlers registered generator function map
var generatorHandlers = map[string]GeneratorHandler{}

// init registers built-in generator functions
func init() {
	Register("flowNo", funcs.FlowNo)
	Register("timestamp", funcs.Timestamp)
	Register("uuid", funcs.UUID)
	Register("random", funcs.Random)
}

// Register registers a dynamic generator function
func Register(name string, fn GeneratorHandler) {
	generatorHandlers[name] = fn
}

// Pre-compiled regex to avoid recompilation on each call
var dynamicRegex = regexp.MustCompile(`\$\(([^,)]+)(?:,(.*?))?\)`)

// Process scans and processes $(funcName,args...) pattern in strings
// Returns the processed string
func Process(value string) string {
	if !strings.Contains(value, "$(") {
		return value
	}

	// Use pre-compiled regex
	return dynamicRegex.ReplaceAllStringFunc(value, func(match string) string {
		parts := dynamicRegex.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}

		funcName := parts[1]
		fn, ok := generatorHandlers[funcName]
		if !ok {
			return match // Function not registered, return original value
		}

		var args []string
		if len(parts) > 2 && parts[2] != "" {
			args = strings.Split(parts[2], ",")
		}

		return fn(args...)
	})
}

// HasDynamicValue detects if value contains dynamic value $(...)
func HasDynamicValue(v any) bool {
	switch val := v.(type) {
	case map[string]any:
		for _, vv := range val {
			if HasDynamicValue(vv) {
				return true
			}
		}
	case []any:
		for _, vv := range val {
			if HasDynamicValue(vv) {
				return true
			}
		}
	case string:
		if strings.Contains(val, "$(") {
			return true
		}
	}
	return false
}

// ProcessValue processes structured response (for HTTP/Dubbo)
// Recursively traverses map[string]any and []any, calls Process() for string types
func ProcessValue(data any, copy bool) any {
	if data == nil {
		return nil
	}

	// Deep copy
	if copy {
		data = DeepCopy(data)
	}
	return processValue(data)
}

// processValue recursively processes dynamic values
func processValue(data any) any {
	switch v := data.(type) {
	case map[string]any:
		return processMap(v)
	case []any:
		return processSlice(v)
	case string:
		return Process(v)
	default:
		return v
	}
}

func processMap(m map[string]any) map[string]any {
	result := make(map[string]any, len(m))
	for k, v := range m {
		result[k] = processValue(v)
	}
	return result
}

func processSlice(s []any) []any {
	result := make([]any, len(s))
	for i, v := range s {
		result[i] = processValue(v)
	}
	return result
}
