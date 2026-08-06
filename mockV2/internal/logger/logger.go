package logger

import (
	"io"
	"os"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	gostlogger "github.com/dubbogo/gost/log/logger"
)

var asyncWriter *AsyncLogWriter

func init() {
	// Create logs directory if not exists
	if err := os.MkdirAll("log", 0755); err != nil {
		panic(err)
	}

	// log scrolling
	logFile := "log/mock.log"

	logWriter, err := rotatelogs.New(
		logFile+".%Y%m%d",
		rotatelogs.WithLinkName(logFile),
		rotatelogs.WithMaxAge(30*24*time.Hour),
		rotatelogs.WithRotationTime(24*time.Hour),
	)

	if err != nil {
		panic(err)
	}

	// print log to file and stdout
	logWriters := io.MultiWriter(os.Stdout, logWriter)

	// Use async log writer
	asyncWriter = NewAsyncLogWriter(logWriters)

	// new logger
	logger := zerolog.New(asyncWriter).
		Level(zerolog.InfoLevel).
		With().
		Timestamp().Logger()

	// reset zerolog logger
	log.Logger = logger

	// Set dubbo-go's logger to zerolog adapter
	gostlogger.SetLogger(NewZerologAdapter(logger))
}

// Close closes the async log writer and waits for pending logs to be written
func Close() error {
	if asyncWriter != nil {
		return asyncWriter.Close()
	}
	return nil
}
