package middleware

import (
	"mock/pkg"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

// RequestLoggingMiddleware HTTP Mock request logging middleware
func RequestLoggingMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		// Generate TraceID
		traceID := pkg.GenerateTraceID()
		c.Locals("traceID", traceID)

		// Process request
		err := c.Next()

		delay := -1
		if d := c.Locals("delay"); d != nil {
			delay = d.(int)
		}

		log.Info().
			Str("traceId", traceID).
			Str("protocol", "http").
			Str("method", c.Method()).
			Str("path", c.Path()).
			Str("clientIp", c.IP()).
			Int("delay", delay).
			Int64("duration", time.Since(start).Milliseconds()).
			Int("statusCode", c.Response().StatusCode()).
			Bytes("request", c.Body()).
			Bytes("response", c.Response().Body()).
			Msg("http request completed")
		return err
	}
}

// ProxyLoggingMiddleware HTTP Proxy logging middleware
// Additional backend selection information is recorded
func ProxyLoggingMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		// Generate TraceID
		traceID := pkg.GenerateTraceID()
		c.Locals("traceID", traceID)

		// Process request
		err := c.Next()

		log.Info().
			Str("traceId", traceID).
			Str("protocol", "http").
			Str("method", c.Method()).
			Str("path", c.Path()).
			Str("clientIp", c.IP()).
			Str("backend", c.Locals("backend").(string)).
			Int("delay", c.Locals("delay").(int)).
			Int64("duration", time.Since(start).Milliseconds()).
			Int("statusCode", c.Response().StatusCode()).
			Bytes("request", c.Body()).
			Bytes("response", c.Response().Body()).
			Msg("http proxy completed")
		return err
	}
}
