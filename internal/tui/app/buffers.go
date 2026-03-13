package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/mdjarv/db/internal/tui/core"
)

const defaultMaxBuffers = 10

// Buffer holds the state of a single query buffer.
type Buffer struct {
	Query    string
	Modified bool

	// result state
	Columns  []core.ResultColumn
	Rows     [][]string
	Duration time.Duration
	HasData  bool
	ErrMsg   string

	// result table scroll positions
	CursorRow int
	CursorCol int
	RowOffset int
	ColOffset int
}

// BufferManager manages multiple query buffers.
type BufferManager struct {
	buffers []*Buffer
	active  int
	max     int
}

// NewBufferManager creates a manager with one empty buffer.
func NewBufferManager() *BufferManager {
	bm := &BufferManager{
		max: defaultMaxBuffers,
	}
	bm.buffers = []*Buffer{{}}
	return bm
}

// Active returns the current buffer.
func (bm *BufferManager) Active() *Buffer {
	return bm.buffers[bm.active]
}

// ActiveIndex returns the 1-based index of the active buffer.
func (bm *BufferManager) ActiveIndex() int {
	return bm.active + 1
}

// Count returns the total number of buffers.
func (bm *BufferManager) Count() int {
	return len(bm.buffers)
}

// New creates a new buffer after the current one and switches to it.
// Returns false if max buffers reached.
func (bm *BufferManager) New() bool {
	if len(bm.buffers) >= bm.max {
		return false
	}
	buf := &Buffer{}
	// insert after active
	pos := bm.active + 1
	bm.buffers = append(bm.buffers, nil)
	copy(bm.buffers[pos+1:], bm.buffers[pos:])
	bm.buffers[pos] = buf
	bm.active = pos
	return true
}

// Close removes the active buffer. Returns false if it's the last buffer.
func (bm *BufferManager) Close() bool {
	if len(bm.buffers) <= 1 {
		return false
	}
	bm.buffers = append(bm.buffers[:bm.active], bm.buffers[bm.active+1:]...)
	if bm.active >= len(bm.buffers) {
		bm.active = len(bm.buffers) - 1
	}
	return true
}

// Next switches to the next buffer, wrapping around.
func (bm *BufferManager) Next() {
	bm.active = (bm.active + 1) % len(bm.buffers)
}

// Prev switches to the previous buffer, wrapping around.
func (bm *BufferManager) Prev() {
	bm.active = (bm.active - 1 + len(bm.buffers)) % len(bm.buffers)
}

// SwitchTo switches to buffer n (1-based). Returns false if out of range.
func (bm *BufferManager) SwitchTo(n int) bool {
	idx := n - 1
	if idx < 0 || idx >= len(bm.buffers) {
		return false
	}
	bm.active = idx
	return true
}

// List returns a formatted string of all buffers.
func (bm *BufferManager) List() string {
	var sb strings.Builder
	for i, b := range bm.buffers {
		marker := " "
		if i == bm.active {
			marker = "%"
		}
		mod := " "
		if b.Modified {
			mod = "+"
		}
		query := b.Query
		if len(query) > 40 {
			query = query[:40] + "..."
		}
		if query == "" {
			query = "[empty]"
		}
		fmt.Fprintf(&sb, " %s%d %s %s\n", marker, i+1, mod, query)
	}
	return sb.String()
}

// Buffers returns the slice of all buffers (for iteration).
func (bm *BufferManager) Buffers() []*Buffer {
	return bm.buffers
}
