package buffer

// EditBuffer manages a single edit buffer datastructure that
// implements Buffer interface{}.
//
// dot - unicode aligned cursor within the buffer starting from 0,
// where a value of N means there are N bytes before the
// cursor, 0 means start and z len(buffer) means.
//
// Not thread safe.
type EditBuffer struct {
	dot    int64  // cursor within the edit buffer
	buffer Buffer // buffer data-structure
	ronly  bool   // buffer is read-only
	// parent and children are used to manage the change tree.
	parent   *EditBuffer
	children []*EditBuffer
}

// NewEditBuffer create a new read-write buffer.
func NewEditBuffer(dot int64, buffer Buffer, parent *EditBuffer) *EditBuffer {
	ebuf := &EditBuffer{
		dot:      dot,
		buffer:   buffer,
		ronly:    false,
		parent:   parent,
		children: make([]*EditBuffer, 0),
	}
	return ebuf
}

// NewReadOnlyBuffer create a new read-only buffer.
func NewReadOnlyBuffer(dot int64, buffer Buffer) *EditBuffer {
	ebuf := NewEditBuffer(dot, buffer, nil)
	ebuf.ronly = true
	return ebuf
}

// GetBuffer return buffer and cursor position.
func (ebuf *EditBuffer) GetBuffer() (dot int64, buffer Buffer) {
	dot, buffer = ebuf.dot, ebuf.buffer
	return
}

// IsReadonly check whether edit-buffer is read-only.
func (ebuf *EditBuffer) IsReadonly() bool {
	return ebuf.ronly
}

// ForceWrite mark edit-buffer as read-write buffer.
func (ebuf *EditBuffer) ForceWrite() *EditBuffer {
	ebuf.ronly = false
	return ebuf
}

//---------------------------
// APIs to manage change-tree
//---------------------------

// UpdateChange will overwrite the current buffer reference.
func (ebuf *EditBuffer) UpdateChange(buffer Buffer) (*EditBuffer, error) {
	if ebuf.ronly {
		return ebuf, ErrorReadonlyBuffer
	}
	ebuf.buffer = buffer
	return ebuf, nil
}

// AppendChange will create a new edit buffer with {dot,buffer}
// and chain it with the current edit buffer as its last child.
func (ebuf *EditBuffer) AppendChange(
	dot int64, buffer Buffer) (*EditBuffer, error) {

	if ebuf.ronly {
		return ebuf, ErrorReadonlyBuffer
	}
	child := NewEditBuffer(dot, buffer, ebuf)
	ebuf.children = append(ebuf.children, child)
	return child, nil
}

// Undo n changes.
func (ebuf *EditBuffer) UndoChange(n int64) *EditBuffer {
	for ebuf.parent != nil && n > 0 {
		ebuf = ebuf.parent
		n--
	}
	return ebuf
}

// Redo n changes.
func (ebuf *EditBuffer) RedoChange(n int64) *EditBuffer {
	for len(ebuf.children) > 0 && n > 0 {
		ebuf = ebuf.children[len(ebuf.children)-1]
		n--
	}
	return ebuf
}

//----------------
// Cursor movement
//----------------

// MoveTo new cursor position to the right, if `dot` is positive,
// else if negative move to left.
func (ebuf *EditBuffer) MoveTo(dot int64, mode int) *EditBuffer {
	ebuf.dot = dot
	return ebuf
}

// RuboutChar before the current-cursor position.
func (ebuf *EditBuffer) RuboutChar(mode byte) (*EditBuffer, error) {
	if ebuf.dot <= 0 {
		return ebuf, nil
	}

	dot := ebuf.dot
	if mode == ModeInsert {
		if buffer, err := ebuf.buffer.DeleteIn(dot-1, 1); err != nil {
			return ebuf, err
		} else if ebuf, err := ebuf.UpdateChange(buffer); err != nil {
			return ebuf, err
		}
		ebuf.dot -= 1
	}
	if buffer, err := ebuf.buffer.Delete(dot-1, 1); err != nil {
		return ebuf, err
	} else if ebuf, err := ebuf.AppendChange(dot-1, buffer); err != nil {
		return ebuf, err
	}
	return ebuf, nil
}

// RuboutWord at current-cursor, if cursor is pointing to
// white-space rubout previous word.
func (ebuf *EditBuffer) RuboutWord() (*EditBuffer, error) {
	return ebuf, nil
}

// RuboutLine at current-cursor including the lines
// end (aka newline char).
func (ebuf *EditBuffer) RuboutLine() (*EditBuffer, error) {
	return ebuf, nil
}
