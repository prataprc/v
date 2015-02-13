package buffer

import "fmt"
import "io"
import "unicode/utf8"

var _ = fmt.Sprintf("dummy")

type LinearRuneReader struct {
	lb *LinearBuffer
}

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
		return nil, ErrorBufferNil
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
	if ch == utf8.RuneError {
		return ch, 0, ErrorInvalidEncoding
	}
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
		return nil, 0, ErrorBufferNil
	} else if rn == 0 {
		return []rune{}, 0, nil
	} else if l := int64(len(lb.Text)); bCur < 0 || bCur >= l {
		return nil, 0, ErrorIndexOutofbound
	} else if l == 0 {
		return nil, 0, ErrorIndexOutofbound
	}
	runes, size := make([]rune, 0, rn), int64(0)
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
		return lb, ErrorBufferNil
	} else if bCur < 0 || bCur > int64(len(lb.Text)) {
		return lb, ErrorIndexOutofbound
	}

	left, right, err := lb.Split(bCur)
	if err != nil {
		return lb, err
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

// Delete implement Buffer{} interface.
func (lb *LinearBuffer) Delete(bCur, rn int64) (*LinearBuffer, error) {
	if lb == nil {
		return lb, ErrorBufferNil
	} else if rn == 0 {
		return lb, nil
	} else if bCur < 0 || bCur > int64(len(lb.Text)-1) {
		return lb, ErrorIndexOutofbound
	}
	runes, size, err := lb.RuneSlice(bCur, rn)
	if err != nil {
		return lb, err
	} else if int64(len(runes)) != rn {
		return lb, ErrorIndexOutofbound
	} else if end := bCur + size; end < 0 || end > int64(len(lb.Text)) {
		return lb, ErrorIndexOutofbound
	}
	newt := make([]byte, int64(len(lb.Text))-size)
	copy(newt[:bCur], lb.Text[:bCur])
	copy(newt[bCur:], lb.Text[bCur+size:])
	newlb := NewLinearBuffer(newt)
	return newlb, nil
}

// InsertIn implement Buffer{} interface.
func (lb *LinearBuffer) InsertIn(bCur int64, text []rune) (*LinearBuffer, error) {
	textb := []byte(string(text)) // TODO: this could be inefficient
	x := int64(len(lb.Text))
	if text == nil || len(textb) == 0 {
		return lb, nil
	} else if lb == nil {
		return lb, ErrorBufferNil
	} else if x == 0 {
		return NewLinearBuffer(textb), nil
	} else if bCur < 0 || bCur > x {
		return lb, ErrorIndexOutofbound
	}

	leftSl := make([]byte, x-bCur)
	copy(leftSl, lb.Text[bCur:])
	lb.Text = append(lb.Text[:bCur], textb...)
	lb.Text = append(lb.Text, leftSl...)
	return lb, nil
}

// DeleteIn implement Buffer{} interface.
func (lb *LinearBuffer) DeleteIn(bCur, rn int64) (*LinearBuffer, error) {
	x := int64(len(lb.Text))
	if lb == nil {
		return lb, ErrorBufferNil
	} else if rn == 0 {
		return lb, nil
	} else if bCur < 0 || bCur > (x-1) {
		return lb, ErrorIndexOutofbound
	}
	_, size, err := lb.RuneSlice(bCur, rn)
	if err != nil {
		return lb, err
	} else if end := bCur + size; end < 0 || end > x {
		return lb, ErrorIndexOutofbound
	}
	copy(lb.Text[bCur:], lb.Text[bCur+size:])
	lb.Text = lb.Text[:x-size]
	return lb, nil
}

// StreamFrom implement Buffer interface{}.
func (lb *LinearBuffer) StreamFrom(bCur int64) io.RuneReader {
	return iterator(func() (r rune, size int, err error) {
		if bCur >= int64(len(lb.Text)) {
			return r, size, io.EOF
		}
		r, size = utf8.DecodeRune(lb.Text[bCur:])
		bCur += int64(size)
		return r, size, nil
	})
}

// StreamTill implement Buffer interface{}.
func (lb *LinearBuffer) StreamTill(bCur, end int64) io.RuneReader {
	return iterator(func() (r rune, size int, err error) {
		if bCur > int64(len(lb.Text)) || bCur >= end {
			return r, size, io.EOF
		}
		r, size = utf8.DecodeRune(lb.Text[bCur:])
		bCur += int64(size)
		return r, size, nil
	})
}

// BackStreamFrom implement Buffer interface{}.
func (lb *LinearBuffer) BackStreamFrom(bCur int64) io.RuneReader {
	return iterator(func() (r rune, size int, err error) {
		if bCur == 0 {
			return r, size, io.EOF
		}
		from := bCur - MaxRuneWidth
		if from < 0 {
			from = 0
		}
		n, err := getRuneStart(lb.Text[from:bCur], true /*reverse*/)
		if err != nil {
			return r, size, err
		}
		bCur = from + n
		r, size = utf8.DecodeRune(lb.Text[bCur:])
		return r, size, nil
	})
}

// BackStreamTill implement Buffer interface{}.
func (lb *LinearBuffer) BackStreamTill(bCur, end int64) io.RuneReader {
	return iterator(func() (r rune, size int, err error) {
		if bCur <= end {
			return r, size, io.EOF
		}
		from := bCur - MaxRuneWidth
		if from < 0 {
			from = 0
		} else if from < end {
			from = end
		}
		n, err := getRuneStart(lb.Text[from:bCur], true /*reverse*/)
		if err != nil {
			return r, size, err
		}
		bCur = from + n
		r, size = utf8.DecodeRune(lb.Text[bCur:])
		return r, size, nil
	})
}

// Stats implement Buffer{} interface.
func (lb *LinearBuffer) Stats() (stats Statistics, err error) {
	return
}
