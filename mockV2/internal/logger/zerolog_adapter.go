package logger

import (
	"github.com/rs/zerolog"
)

// ZerologAdapter zerolog adapter, implements gost logger interface
type ZerologAdapter struct {
	l zerolog.Logger
}

// NewZerologAdapter creates zerolog adapter
func NewZerologAdapter(l zerolog.Logger) *ZerologAdapter {
	return &ZerologAdapter{l: l}
}

// Info implements Logger interface
func (a *ZerologAdapter) Info(args ...interface{}) {
	a.l.Info().Msg(formatArgs(args))
}

// Warn implements Logger interface
func (a *ZerologAdapter) Warn(args ...interface{}) {
	a.l.Warn().Msg(formatArgs(args))
}

// Error implements Logger interface
func (a *ZerologAdapter) Error(args ...interface{}) {
	a.l.Error().Msg(formatArgs(args))
}

// Debug implements Logger interface
func (a *ZerologAdapter) Debug(args ...interface{}) {
	a.l.Debug().Msg(formatArgs(args))
}

// Fatal implements Logger interface
func (a *ZerologAdapter) Fatal(args ...interface{}) {
	a.l.Fatal().Msg(formatArgs(args))
}

// Infof implements Logger interface
func (a *ZerologAdapter) Infof(fmt string, args ...interface{}) {
	a.l.Info().Msgf(fmt, args...)
}

// Warnf implements Logger interface
func (a *ZerologAdapter) Warnf(fmt string, args ...interface{}) {
	a.l.Warn().Msgf(fmt, args...)
}

// Errorf implements Logger interface
func (a *ZerologAdapter) Errorf(fmt string, args ...interface{}) {
	a.l.Error().Msgf(fmt, args...)
}

// Debugf implements Logger interface
func (a *ZerologAdapter) Debugf(fmt string, args ...interface{}) {
	a.l.Debug().Msgf(fmt, args...)
}

// Fatalf implements Logger interface
func (a *ZerologAdapter) Fatalf(fmt string, args ...interface{}) {
	a.l.Fatal().Msgf(fmt, args...)
}

// formatArgs formats args list to string
func formatArgs(args []interface{}) string {
	result := ""
	for i, arg := range args {
		if i > 0 {
			result += " "
		}
		result += toString(arg)
	}
	return result
}

// toString converts any type to string
func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case error:
		return val.Error()
	default:
		return toString(val)
	}
}
