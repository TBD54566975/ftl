package exec

import (
	"bufio"
	"bytes"
	"container/ring"
	"context"
	"fmt"
	"io"
	"runtime"
	"sync"

	"github.com/TBD54566975/ftl/backend/common/log"
)

type CircularBuffer struct {
	r    *ring.Ring
	size int
	mu   sync.Mutex
	cap  int
}

func NewCircularBuffer(capacity int) *CircularBuffer {
	return &CircularBuffer{
		r:   ring.New(capacity),
		cap: capacity,
	}
}

// Write accepts a string and stores it in the buffer.
// It expects entire lines and stores each line as a separate entry in the ring buffer.
func (cb *CircularBuffer) Write(p string) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.r.Value = p
	cb.r = cb.r.Next()

	if cb.size < cb.cap {
		cb.size++
	}

	return nil
}

func (cb *CircularBuffer) Bytes() []byte {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.size == 0 {
		fmt.Println("Buffer is empty.")
		return []byte{}
	}

	var buf bytes.Buffer
	start := cb.r.Move(-cb.size) // Correctly calculate the starting position

	for i := 0; i < cb.size; i++ {
		if str, ok := start.Value.(string); ok {
			buf.WriteString(str)
		} else {
			fmt.Println("Unexpected type or nil found in buffer")
		}
		start = start.Next()
	}

	return buf.Bytes()
}

func (cb *CircularBuffer) WriterAt(ctx context.Context, level log.Level) *io.PipeWriter {
	// Copied from logger.go which is based on logrus
	// Based on MIT licensed Logrus https://github.com/sirupsen/logrus/blob/bdc0db8ead3853c56b7cd1ac2ba4e11b47d7da6b/writer.go#L27
	logger := log.FromContext(ctx)
	reader, writer := io.Pipe()
	var printFunc func(format string, args ...interface{})

	switch level {
	case log.Trace:
		printFunc = logger.Tracef
	case log.Debug:
		printFunc = logger.Debugf
	case log.Info:
		printFunc = logger.Infof
	case log.Warn:
		printFunc = logger.Warnf
	case log.Error:
		printFunc = func(format string, args ...interface{}) {
			logger.Errorf(nil, format, args...)
		}
	default:
		panic(level)
	}

	go cb.writerScanner(reader, printFunc)
	runtime.SetFinalizer(writer, writerFinalizer)

	return writer
}

func (cb *CircularBuffer) writerScanner(reader *io.PipeReader, printFunc func(format string, args ...interface{})) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		text := scanner.Text()
		printFunc(text)
		err := cb.Write(text + "\n")
		if err != nil {
			fmt.Println("Error writing to buffer")
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading from pipe:", err)
	}
	reader.Close()
}

func writerFinalizer(writer *io.PipeWriter) {
	writer.Close()
}
