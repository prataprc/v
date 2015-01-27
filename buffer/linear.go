package buffer

import "fmt"
import "unicode/utf8"

var _ = fmt.Sprintf("dummy")

// LinearBuffer represents mutable sequence of runes as buffer.
type LinearBuffer struct {
	Text []byte
}

// NewLinearBuffer returns a new buffer, initialized with text.
func NewLinearBuffer(text []byte) *LinearBuffer {
	newt := make([]byte, len(text))
	copy(newt, text)
	return &LinearBuffer{Text: newt}
}

// Length implement Buffer{} interface.
func (lb *LinearBuffer) Length() (n int64, err error) {
	if lb == nil {
		return n, ErrorBufferNil
	}
	return int64(len(lb.Text)), nil
}

// Value implement Buffer{} interface.
func (lb *LinearBuffer) Value() []byte {
	if lb == nil {
		return nil
	}
	return lb.Text
}

func (lb *LinearBuffer) Slice(bCur, bn int64) ([]byte, error) {
	if lb == nil {
		return nil, nil
	} else if l := int64(len(lb.Text)); bCur < 0 || bCur > l {
		return nil, ErrorIndexOutofbound
	} else if end := bCur + bn; end < 0 || end > int64(len(lb.Text)) {
		return nil, ErrorIndexOutofbound
	}
	return lb.Text[bCur : bCur+bn], nil
}

// RuneAt implement Buffer{} interface.
func (lb *LinearBuffer) RuneAt(bCur int64) (ch rune, size int, err error) {
	if lb == nil {
		return ch, size, ErrorBufferNil
	} else if l := int64(len(lb.Text)); bCur < 0 || bCur >= l {
		return ch, size, ErrorIndexOutofbound
	}
	ch, size = utf8.DecodeRune(lb.Text[bCur:])
	return ch, size, nil
}

// Runes implement Buffer{} interface.
func (lb *LinearBuffer) Runes() ([]rune, error) {
	if lb == nil {
		return nil, ErrorBufferNil
	}
	return bytes2Runes(lb.Text)
}

// RuneSlice implement Buffer{} interface.
func (lb *LinearBuffer) RuneSlice(bCur, rn int64) ([]rune, int64, error) {
	if lb == nil {
		return nil, 0, nil
	} else if rn == 0 {
		return []rune{}, 0, nil
	} else if l := int64(len(lb.Text)); bCur < 0 || bCur >= l {
		return nil, 0, ErrorIndexOutofbound
	} else if l == 0 {
		return nil, 0, ErrorIndexOutofbound
	}
	runes, size := make([]rune, rn), int64(0)
	for i := int64(0); i < rn; i++ {
		ch, sz := utf8.DecodeRune(lb.Text[bCur+size:])
		if ch == utf8.RuneError {
			return nil, 0, ErrorInvalidEncoding
		}
		runes = append(runes, ch)
		size += int64(sz)
		if bCur+size >= int64(len(lb.Text)) {
			break
		}
	}
	return runes, size, nil
}

// Concat implement Buffer{} interface.
func (lb *LinearBuffer) Concat(right *LinearBuffer) (*LinearBuffer, error) {
	if lb == nil {
		return right, nil
	} else if right == nil {
		return lb, nil
	}
	newt := make([]byte, len(lb.Text)+len(right.Text))
	copy(newt, lb.Text)
	copy(newt[len(lb.Text):], right.Text)
	newlb := NewLinearBuffer(newt)
	return newlb, nil
}

// Split implement Buffer{} interface.
func (lb *LinearBuffer) Split(bCur int64) (left, right *LinearBuffer, err error) {
	if lb == nil {
		return left, right, ErrorBufferNil
	} else if bCur < 0 || bCur > int64(len(lb.Text)) {
		return left, right, ErrorIndexOutofbound
	} else if bCur == 0 {
		return nil, lb, nil
	} else if bCur == int64(len(lb.Text)) {
		return lb, nil, nil
	}
	lsize, rsize := bCur, int64(len(lb.Text))-bCur
	l, r := make([]byte, lsize), make([]byte, rsize)
	copy(l, lb.Text[:lsize])
	copy(r, lb.Text[lsize:])
	return NewLinearBuffer(l), NewLinearBuffer(r), nil
}

// Insert implement Buffer{} interface.
func (lb *LinearBuffer) Insert(bCur int64, text []rune) (*LinearBuffer, error) {
	textb := []byte(string(text)) // TODO: this could be inefficient.
	if text == nil {
		return lb, nil
	} else if lb == nil {
		return NewLinearBuffer(textb), nil
	}

	left, right, err := lb.Split(bCur)
	if err != nil {
		return nil, err
	}
	if left == nil {
		left = NewLinearBuffer([]byte(""))
	}
	if right == nil {
		right = NewLinearBuffer([]byte(""))
	}
	newlb := make([]byte, len(lb.Text)+len(textb))
	copy(newlb, left.Text)
	copy(newlb[len(left.Text):], textb)
	copy(newlb[len(left.Text)+len(textb):], right.Text)
	return NewLinearBuffer(newlb), nil
}

// InsertIO implement Buffer{} interface.
func (lb *LinearBuffer) InsertIO(bCur int64, text []rune) (*LinearBuffer, error) {
	return lb.Insert(bCur, text)
}

// Delete implement Buffer{} interface.
func (lb *LinearBuffer) Delete(bCur, Rn int64) (*LinearBuffer, error) {
	if lb == nil {
		return nil, ErrorBufferNil
	} else if bCur < 0 || bCur > int64(len(lb.Text)-1) {
		return nil, ErrorIndexOutofbound
	}
	_, size, err := lb.RuneSlice(bCur, Rn)
	if err != nil {
		return nil, err
	} else if end := bCur + size; end < 0 || end > int64(len(lb.Text)) {
		return nil, ErrorIndexOutofbound
	}
	newt := make([]byte, int64(len(lb.Text))-size)
	copy(newt[:bCur], lb.Text[:bCur])
	copy(newt[bCur:], lb.Text[bCur+size:])
	newlb := NewLinearBuffer(newt)
	return newlb, nil
}

// DeleteIO implement Buffer{} interface.
func (lb *LinearBuffer) DeleteIO(bCur, n int64) (*LinearBuffer, error) {
	return lb.Delete(bCur, n)
}

// Stats implement Buffer{} interface.
func (lb *LinearBuffer) Stats() (stats Statistics, err error) {
	return
}
