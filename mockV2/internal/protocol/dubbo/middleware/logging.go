package middleware

import (
	"time"

	"dubbo.apache.org/dubbo-go/v3/common/constant"
	"dubbo.apache.org/dubbo-go/v3/protocol"
	"dubbo.apache.org/dubbo-go/v3/protocol/invocation"
	"github.com/rs/zerolog/log"
	"mock/pkg"
)

// DubboHandler Dubbo request handler function type
type DubboHandler func(*invocation.RPCInvocation) protocol.RPCResult

// RequestLoggingMiddleware creates Dubbo logging middleware
// Returns a decorator function that takes the original handler and returns a wrapped handler
func RequestLoggingMiddleware() func(DubboHandler) DubboHandler {
	return func(handler DubboHandler) DubboHandler {
		return func(in *invocation.RPCInvocation) protocol.RPCResult {
			start := time.Now()
			traceID := pkg.GenerateTraceID()

			// Call the original handler to process the request
			result := handler(in)

			log.Info().
				Str("traceId", traceID).
				Str("protocol", "dubbo").
				Str("method", in.MethodName()).
				Str("path", func() string {
					if path, ok := in.Attachments()[constant.PathKey].(string); ok {
						return path
					}
					return ""
				}()).
				Int64("duration", time.Since(start).Milliseconds()).
				Interface("request", map[string]any{
					"arguments":  in.Arguments(),
					"paramTypes": in.ParameterTypeNames(),
				}).
				Interface("response", result.Result()).
				Msg("dubbo request completed")

			return result
		}
	}
}
