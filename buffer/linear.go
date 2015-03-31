package buffer

import "fmt"
import "io"
import "unicode/utf8"

var _ = fmt.Sprintf("dummy")

// LinearBuffer implements a immutable array of bytes as
// Buffer.
type LinearBuffer struct {
	Text []rune
}

// NewLinearBuffer returns a new buffer,
// initialized with text, by copying it to locally.
func NewLinearBuffer(bs []byte) Buffer {
	return &LinearBuffer{Text: bytes2Runes(bs)}
}

//----------
// rune APIs
//----------

// Length implement Buffer{} interface.
func (lb *LinearBuffer) Length() int64 {
	return int64(len(lb.Text))
}

// Slice implement Buffer{} interface.
func (lb *LinearBuffer) Slice(rCur, rn int64) Buffer {
	endCur := rCur + rn
	if !lb.isValidCursor(rCur) {
		return nil
	} else if ok := lb.isValidCursor(endCur); !ok && endCur < rCur {
		return nil
	} else if !ok {
		endCur = lb.Length()
	}
	rs := make([]rune, endCur-rCur)
	copy(rs, lb.Text[rCur:endCur])
	return &LinearBuffer{Text: rs}
}

// Runes implement Buffer{} interface.
func (lb *LinearBuffer) Runes() []rune {
	if len(lb.Text) == 0 {
		return nil
	}
	rs := make([]rune, lb.Length())
	copy(rs, lb.Text)
	return rs
}

// Concat implement Buffer{} interface.
func (lb *LinearBuffer) Concat(right Buffer) Buffer {
	rlb := right.(*LinearBuffer)
	if lb == nil {
		return rlb
	} else if rlb == nil {
		return lb
	}
	rs := make([]rune, len(lb.Text)+len(rlb.Text))
	copy(rs, lb.Text)
	copy(rs[len(lb.Text):], rlb.Text)
	return &LinearBuffer{Text: rs}
}

// Split implement Buffer{} interface.
func (lb *LinearBuffer) Split(rCur int64) (left, right Buffer) {
	l := lb.Length()
	if lb == nil {
		return left, right
	} else if rCur >= l {
		return lb, nil
	} else if rCur == 0 {
		return nil, lb
	}
	left = lb.Slice(0, rCur)
	right = lb.Slice(rCur, l-rCur)
	return
}

// Insert implement Buffer{} interface.
func (lb *LinearBuffer) Insert(rCur int64, text []rune) Buffer {
	if text == nil {
		return lb
	} else if lb == nil {
		panic(ErrorBufferNil)
	} else if !lb.isValidCursor(rCur) {
		panic(ErrorIndexOutofbound)
	}
	currLen, insLen := lb.Length(), int64(len(text))
	rs := make([]rune, currLen+insLen)
	copy(rs, lb.Text[:rCur])
	copy(rs[rCur:], text)
	copy(rs[rCur+insLen:], lb.Text[rCur:])
	return &LinearBuffer{Text: rs}
}

// Delete implement Buffer{} interface.
func (lb *LinearBuffer) Delete(rCur, rn int64) Buffer {
	if rn == 0 {
		return lb
	} else if lb == nil {
		panic(ErrorBufferNil)
	} else if !lb.isValidCursor(rCur) {
		panic(ErrorIndexOutofbound)
	}
	rs := make([]rune, int64(len(lb.Text))-rn)
	copy(rs[:rCur], lb.Text[:rCur])
	copy(rs[rCur:], lb.Text[rCur+rn:])
	return &LinearBuffer{Text: rs}
}

// InsertIn implement Buffer{} interface.
func (lb *LinearBuffer) InsertIn(rCur int64, text []rune) Buffer {
	if text == nil {
		return lb
	} else if lb == nil {
		panic(ErrorBufferNil)
	} else if !lb.isValidCursor(rCur) {
		panic(ErrorIndexOutofbound)
	}
	l := int64(len(text))
	lb.Text = append(lb.Text, lb.Text[rCur:]...)
	copy(lb.Text[rCur:rCur+l], text)
	return lb
}

// DeleteIn implement Buffer{} interface.
func (lb *LinearBuffer) DeleteIn(rCur, rn int64) Buffer {
	if rn == 0 {
		return lb
	} else if lb == nil {
		panic(ErrorBufferNil)
	} else if !lb.isValidCursor(rCur) {
		panic(ErrorIndexOutofbound)
	}
	l := int64(len(lb.Text))
	copy(lb.Text[rCur:], lb.Text[rCur+rn:])
	lb.Text = lb.Text[:l-rn]
	return lb
}

//------------
// search APIs
//------------

// StreamFrom implement Buffer interface{}.
func (lb *LinearBuffer) StreamFrom(rCur int64) RuneReader {
	if !lb.isValidCursor(rCur) {
		return nil
	}
	ln := int64(len(lb.Text))
	return iterator(func(finish bool) (r rune, size int, err error) {
		if rCur >= ln || finish {
			return r, size, io.EOF
		}
		r = lb.Text[rCur]
		rCur++
		return r, utf8.RuneLen(r), nil
	})
}

// StreamCount implement Buffer interface{}.
func (lb *LinearBuffer) StreamCount(rCur, count int64) RuneReader {
	if !lb.isValidCursor(rCur) {
		return nil
	}
	ln := int64(len(lb.Text))
	return iterator(func(finish bool) (r rune, size int, err error) {
		if rCur >= ln || count <= 0 || finish {
			return r, size, io.EOF
		}
		r = lb.Text[rCur]
		count--
		rCur++
		return r, utf8.RuneLen(r), nil
	})
}

// BackStreamFrom implement Buffer interface{}.
func (lb *LinearBuffer) BackStreamFrom(rCur int64) RuneReader {
	if !lb.isValidCursor(rCur) {
		return nil
	}
	return iterator(func(finish bool) (r rune, size int, err error) {
		rCur--
		if rCur <= 0 || finish {
			return r, size, io.EOF
		}
		r = lb.Text[rCur]
		return r, utf8.RuneLen(r), nil
	})
}

// BackStreamCount implement Buffer interface{}.
func (lb *LinearBuffer) BackStreamCount(rCur, count int64) RuneReader {
	if !lb.isValidCursor(rCur) {
		return nil
	}
	return iterator(func(finish bool) (r rune, size int, err error) {
		rCur--
		if rCur <= 0 || count <= 0 || finish {
			return r, size, io.EOF
		}
		r = lb.Text[rCur]
		count--
		return r, utf8.RuneLen(r), nil
	})
}

//----------
// byte APIs
//----------

// Size implement Buffer{} interface.
func (lb *LinearBuffer) Size() int64 {
	return int64(len(lb.Bytes()))
}

// Bytes implement Buffer{} interface.
func (lb *LinearBuffer) Bytes() []byte {
	return runes2Bytes(lb.Text)
}

// Stats implement Buffer{} interface.
func (lb *LinearBuffer) Stats() (stats Statistics, err error) {
	return
}

//---------------
// local function
//---------------

func (lb *LinearBuffer) isValidCursor(rCur int64) bool {
	return 0 <= rCur && rCur <= lb.Length()
}
