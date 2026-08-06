package logger

import (
	"io"
	"sync"
)

// Async log writer configuration
const (
	asyncBufferSize = 10000 // channel buffer size
)

// AsyncLogWriter async log writer, implements io.Writer interface
type AsyncLogWriter struct {
	logChan chan []byte // log byte buffer channel
	writer  io.Writer   // target writer
	wg      sync.WaitGroup
	closed  bool
	mu      sync.Mutex
}

// NewAsyncLogWriter creates async log writer
func NewAsyncLogWriter(writer io.Writer) *AsyncLogWriter {
	logWriter := &AsyncLogWriter{
		logChan: make(chan []byte, asyncBufferSize),
		writer:  writer,
	}

	logWriter.wg.Add(1)
	go logWriter.writerLoop()
	return logWriter
}

// Write implements io.Writer interface, writes logs to channel
func (w *AsyncLogWriter) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	w.mu.Lock()
	closed := w.closed
	w.mu.Unlock()

	if closed {
		return len(p), nil
	}

	select {
	case w.logChan <- p:
		return len(p), nil
	default:
		// When channel is full, drop logs without blocking caller
		return len(p), nil
	}
}

// writerLoop worker background log processing goroutine
func (w *AsyncLogWriter) writerLoop() {
	defer w.wg.Done()
	for p := range w.logChan {
		_, _ = w.writer.Write(p)
	}
}

// Close closes the async writer and waits for pending logs to be written
func (w *AsyncLogWriter) Close() error {
	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		return nil
	}
	w.closed = true
	w.mu.Unlock()

	close(w.logChan)
	w.wg.Wait()
	return nil
}
