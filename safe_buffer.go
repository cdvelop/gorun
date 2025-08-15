package gorun

import (
	"bytes"
	"io"
	"sync"
)

// SafeBuffer provides thread-safe operations on a bytes.Buffer
// and optionally forwards writes to another writer
type SafeBuffer struct {
	buffer    *bytes.Buffer
	mutex     sync.RWMutex
	forwardTo io.Writer // Optional writer to forward data to
}

// NewSafeBuffer creates a new thread-safe buffer
func NewSafeBuffer() *SafeBuffer {
	return &SafeBuffer{
		buffer: &bytes.Buffer{},
	}
}

// NewSafeBufferWithForward creates a new thread-safe buffer that forwards writes
func NewSafeBufferWithForward(forward io.Writer) *SafeBuffer {
	// Wrap the forward writer in a thread-safe wrapper so that
	// writes forwarded from SafeBuffer don't race when the original
	// writer is not safe for concurrent use (e.g., bytes.Buffer).
	return &SafeBuffer{
		buffer:    &bytes.Buffer{},
		forwardTo: &forwardWriter{w: forward},
	}
}

// forwardWriter serializes writes to an underlying writer.
// This protects non-thread-safe writers (like bytes.Buffer)
// when SafeBuffer forwards bytes to them from multiple goroutines.
type forwardWriter struct {
	w  io.Writer
	mu sync.Mutex
}

func (fw *forwardWriter) Write(p []byte) (int, error) {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	return fw.w.Write(p)
}

// Write writes data to the buffer and optionally forwards to another writer
func (sb *SafeBuffer) Write(p []byte) (n int, err error) {
	sb.mutex.Lock()
	defer sb.mutex.Unlock()

	// Write to internal buffer first
	n, err = sb.buffer.Write(p)
	if err != nil {
		return n, err
	}

	// Forward to original writer if configured
	if sb.forwardTo != nil {
		_, forwardErr := sb.forwardTo.Write(p)
		// We don't return the forward error since we successfully wrote to our buffer
		// But we could log it if needed
		_ = forwardErr
	}

	return n, err
}

// String returns the contents of the buffer as a string in a thread-safe manner
func (sb *SafeBuffer) String() string {
	sb.mutex.RLock()
	defer sb.mutex.RUnlock()
	return sb.buffer.String()
}

// Reset resets the buffer in a thread-safe manner
func (sb *SafeBuffer) Reset() {
	sb.mutex.Lock()
	defer sb.mutex.Unlock()
	sb.buffer.Reset()
}

// Len returns the length of the buffer in a thread-safe manner
func (sb *SafeBuffer) Len() int {
	sb.mutex.RLock()
	defer sb.mutex.RUnlock()
	return sb.buffer.Len()
}
