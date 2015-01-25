// +build ignore

package buffer

import "sync"
import "errors"

var ErrorReadonlyBuffer = errors.New("editbuffer.ronly")

var ErrorOldestChange = errors.New("editbuffer.oldestChange")

var ErrorLatestChange = errors.New("editbuffer.latestChange")

type EditBuffer struct {
	mu       sync.Mutex
	dot      int64
	buffer   Buffer
	ronly    bool
	parent   *EditBuffer
	children []*EditBuffer
}

func NewEditBuffer(dot uint64, buffer Buffer, parent *EditBuffer) *EditBuffer {
	ebuf := &EditBuffer{
		dot:      0,
		buffer:   buffer,
		ronly:    false,
		parent:   parent,
		children: make([]*EditBuffer, 0),
	}
	return ebuf
}

func NewReadOnlyBuffer(buffer Buffer) *EditBuffer {
	ebuf := &EditBuffer{
		dot:    0,
		buffer: buffer,
		ronly:  true,
	}
	return ebuf
}

func (ebuf *EditBuffer) GetBuffer() (dot int64, buffer Buffer) {
	dot, buffer = ebuf.dot, ebuf.buffer
	return
}

func (ebuf *EditBuffer) IsReadonly() bool {
	return ebuf.ronly
}

//---------------------------
// APIs to manage change-tree
//---------------------------

func (ebuf *EditBuffer) UpdateChange(buffer Buffer) (*EditBuffer, error) {
	if ebuf.ronly {
		return ebuf, ErrorReadonlyBuffer
	}
	ebuf.mu.Lock()
	defer ebuf.mu.Unlock()
	ebuf.buffer = buffer
	return ebuf, nil
}

func (ebuf *EditBuffer) AppendChange(
	dot uint64, buffer Buffer, ronly bool) (*EditBuffer, error) {

	if ebuf.ronly {
		return ebuf, ErrorReadonlyBuffer
	}
	ebuf.mu.Lock()
	defer ebuf.mu.Unlock()
	child := NewEditBuffer(dot, buffer, ebuf)
	child.ronly = ronly
	ebuf.children = append(ebuf.children, child)
	return child, nil
}

func (ebuf *EditBuffer) UndoChange() (*EditBuffer, error) {
	ebuf.mu.Lock()
	defer ebuf.mu.Unlock()
	undo := ebuf.parent
	if undo == nil {
		return ebuf, ErrorOldestChange
	}
	return undo, nil
}

func (ebuf *EditBuffer) RedoChange() (*EditBuffer, error) {
	ebuf.mu.Lock()
	defer ebuf.mu.Unlock()
	l := len(ebuf.children)
	if l < 1 {
		return ebuf, ErrorLatestChange
	}
	return ebuf.children[l-1], nil
}

//-----
// TECO
//-----

// MoveTo new cursor position to the right, if `dot` is positive,
// else if negative move to left.
func (ebuf *EditBuffer) MoveTo(dot int64) (*EditBuffer, error) {
}

// RuboutChar before the current-cursor position.
func (ebuf *EditBuffer) RuboutChar() (*EditBuffer, error) {
}

// RuboutWord at current-cursor, if cursor is pointing to white-space
// rubout previous word.
// - word is delineated by some white-space such as a space,
//   tab or newline.
func (ebuf *EditBuffer) RuboutWord() (*EditBuffer, error) {
}

// RuboutLine at current-cursor including the lines
// end (aka newline char).
func (ebuf *EditBuffer) RuboutLine() (*EditBuffer, error) {
}
