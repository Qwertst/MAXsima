// Package testutil provides shared test helpers for the MAXsima test suite.
package testutil

import (
	"io"
	"sync"

	"github.com/aydreq/maxsima/internal/model"
)

// MockUI is a minimal ui.UIRenderer implementation for tests.
// It records all displayed messages and serves pre-configured input strings.
// When the input queue is exhausted it returns io.EOF.
type MockUI struct {
	mu        sync.Mutex
	Displayed []model.Message
	inputs    []string
	inputIdx  int
}

// DisplayMessage records msg in Displayed.
func (m *MockUI) DisplayMessage(msg model.Message) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Displayed = append(m.Displayed, msg)
}

// ReadInput returns the next queued input string, or io.EOF when exhausted.
func (m *MockUI) ReadInput() (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.inputIdx < len(m.inputs) {
		s := m.inputs[m.inputIdx]
		m.inputIdx++
		return s, nil
	}
	return "", io.EOF
}

// SetInputs replaces the input queue.
func (m *MockUI) SetInputs(inputs []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.inputs = inputs
	m.inputIdx = 0
}

// Messages returns a snapshot of all displayed messages.
func (m *MockUI) Messages() []model.Message {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]model.Message, len(m.Displayed))
	copy(out, m.Displayed)
	return out
}

// BlockingMockUI is a ui.UIRenderer whose ReadInput blocks until Stop() is
// called (at which point it returns io.EOF). Use this for the server side in
// integration tests so the server session stays alive while the test runs.
type BlockingMockUI struct {
	mu        sync.Mutex
	Displayed []model.Message
	stop      chan struct{}
	once      sync.Once
}

// NewBlockingMockUI creates a BlockingMockUI ready for use.
func NewBlockingMockUI() *BlockingMockUI {
	return &BlockingMockUI{stop: make(chan struct{})}
}

// DisplayMessage records msg in Displayed.
func (b *BlockingMockUI) DisplayMessage(msg model.Message) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Displayed = append(b.Displayed, msg)
}

// ReadInput blocks until Stop() is called, then returns io.EOF.
func (b *BlockingMockUI) ReadInput() (string, error) {
	<-b.stop
	return "", io.EOF
}

// Stop unblocks all pending ReadInput calls.
func (b *BlockingMockUI) Stop() {
	b.once.Do(func() { close(b.stop) })
}

// Messages returns a snapshot of all displayed messages.
func (b *BlockingMockUI) Messages() []model.Message {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := make([]model.Message, len(b.Displayed))
	copy(out, b.Displayed)
	return out
}
