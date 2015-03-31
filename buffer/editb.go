// +build ignore

package buffer

import "regexp"
import "fmt"

type Finder func() []int

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
	// buffer settings
	newline string // list of runes that act as newline
	// buffer context
	lines Lines
	atEol bool           // stick cursor to end-of-line
	atBol bool           // stick cursor to beginning-of-line
	reNl  *regexp.Regexp // compiled newline
	reNlR *regexp.Regexp // compiled newline in reverse order
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
	ebuf.Initialize(ebuf)
	return ebuf
}

// NewReadOnlyBuffer create a new read-only buffer.
func NewReadOnlyBuffer(dot int64, buffer Buffer) *EditBuffer {
	ebuf := NewEditBuffer(dot, buffer, nil)
	ebuf.ronly = true
	return ebuf
}

// Initialize EditBuffer.
func (ebuf *EditBuffer) Initialize(parent *EditBuffer) *EditBuffer {
	var err error

	if ebuf.newline == "" {
		ebuf.newline = Newline
	}
	nl := ebuf.newline
	if ebuf.reNl, err = regexp.Compile(nl); err != nil {
		panic("impossible regular expression")
	}
	ebuf.reNlR, err = regexp.Compile(string(reverseRunes([]rune(nl))))
	if err != nil {
		err = fmt.Errorf("impossible regular expression: %v", err)
		panic(err)
	}
	return ebuf
}

// Configure EditBuffer.
func (ebuf *EditBuffer) Configure(setts map[string]interface{}) *EditBuffer {
	ebuf.newline = setts["newline"].(string)
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

// LineAround return a block of consecutive lines around bCur.
// width number of lines above the line containing bCur, and,
// width number of lines below the line containing bcur.
// including the line containing bCur.
func (ebuf *EditBuffer) LinesAround(bCur int64, width int64) Lines {
	block := ebuf.lines.blocksFrom(bCur)()
	if block == nil || (int64(len(block)) < (width*2 + 2)) {
		block := ebuf.BuildBlock(bCur, width)
		ebuf.lines = ebuf.lines.mergeBlock(block)
		return block
	}
	i, ok := block.indexof(bCur)
	if !ok {
		panic("impossible situation\n")
	}
	start := i - width
	if start < 0 {
		start = 0
	}
	end := i + width
	if end >= int64(len(block)) {
		end = int64(len(block)) - 1
	}
	lines := make(Lines, 0, width*2+2)
	for i := start; i <= end; i++ {
		lines = append(lines, block[i])
	}
	return lines
}

// BuildBlock around specified cursor position,
// if `bCur` is -1 use current cursor position.
// Return Lines of specified width*2 + 1.
func (ebuf *EditBuffer) BuildBlock(bCur int64, width int64) Lines {
	if bCur < 0 {
		bCur = ebuf.dot
	}
	lnNl := int64(len(ebuf.newline))
	lines := make(Lines, 0, (width*2+1)*2+2)
	// gather lines above cursor.
	iter := Find(ebuf.reNlR, ebuf.buffer.BackStreamFrom(bCur))
	for i, end := int64(0), int64(-1); i < width+1; i++ {
		loc := iter()
		if loc != nil {
			lines = append(lines, end, int64(loc[0])+lnNl)
			end = int64(loc[1])
		} else {
			lines = append(lines, end, 0)
			break
		}
	}
	lines.reverse()
	// gather lines below cursor.
	iter = Find(ebuf.reNlR, ebuf.buffer.StreamFrom(bCur))
	for i, start := int64(0), int64(-1); i < width; i++ {
		loc := iter()
		if loc != nil {
			lines = append(lines, start, int64(loc[1]))
			start = int64(loc[1])
		} else {
			lines = append(lines, start, ebuf.buffer.Size())
			break
		}
	}
	// consolidate.
	for i, j := 1, 1; (i + 1) < len(lines); i += 2 {
		if lines[i] != lines[i+1] {
			panic("impossible situation")
		} else if lines[i] == -1 && lines[i+1] == -1 {
			continue
		}
		lines[j], lines[j+1] = lines[i], lines[i+1]
		j += 2
	}
	return lines
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
