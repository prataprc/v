package rope

import "fmt"
import "github.com/prataprc/v"

var _ = fmt.Sprintf("dummy")

// LinearBuffer represents mutable sequence of runes as buffer.
type LinearBuffer struct {
	Dot  int64
	Text []rune
}

// NewLinearBuffer returns a new buffer, initialized with text.
func NewLinearBuffer(text []rune) *LinearBuffer {
	newt := make([]rune, len(text))
	copy(newt, text)
	return &LinearBuffer{Dot: 0, Text: newt}
}

// Length implement Buffer{} interface.
func (lb *LinearBuffer) Length() (n int64, err error) {
	if lb == nil {
		return n, v.ErrorBufferNil
	}
	return int64(len(lb.Text)), nil
}

// Value implement Buffer{} interface.
func (lb *LinearBuffer) Value() []rune {
	if lb == nil {
		return nil
	}
	return lb.Text
}

// Index implement Buffer{} interface.
func (lb *LinearBuffer) Index(dot int64) (ch rune, ok bool, err error) {
	if lb == nil {
		return ch, false, v.ErrorBufferNil
	} else if l := int64(len(lb.Text)); dot < 0 || dot > l {
		return ch, false, v.ErrorIndexOutofbound
	} else if dot == l {
		return ch, false, v.ErrorIndexOutofbound
	}
	return lb.Text[dot], true, nil
}

// Substr implement Buffer{} interface.
func (lb *LinearBuffer) Substr(dot int64, n int64) (string, error) {
	if lb == nil {
		return "", nil
	} else if l := int64(len(lb.Text)); dot < 0 || dot > l {
		return "", v.ErrorIndexOutofbound
	} else if dot+n > l {
		return "", v.ErrorIndexOutofbound
	}
	return string(lb.Text[dot : dot+n]), nil
}

// Concat implement Buffer{} interface.
func (lb *LinearBuffer) Concat(right *LinearBuffer) (*LinearBuffer, error) {
	if lb == nil {
		return right, nil
	} else if right == nil {
		return lb, nil
	}
	newt := make([]rune, len(lb.Text)+len(right.Text))
	copy(newt, lb.Text)
	copy(newt[len(lb.Text):], right.Text)
	newlb := NewLinearBuffer(newt)
	return newlb, nil
}

// Split implement Buffer{} interface.
func (lb *LinearBuffer) Split(dot int64) (left, right *LinearBuffer, err error) {
	if lb == nil {
		return left, right, v.ErrorBufferNil
	} else if dot < 0 || dot > int64(len(lb.Text)) {
		return left, right, v.ErrorIndexOutofbound
	} else if dot == 0 {
		return nil, lb, nil
	} else if dot == int64(len(lb.Text)) {
		return lb, nil, nil
	}
	l, r := make([]rune, dot), make([]rune, int64(len(lb.Text))-dot)
	copy(l, lb.Text[:dot])
	copy(r, lb.Text[dot:])
	return NewLinearBuffer(l), NewLinearBuffer(r), nil
}

// Insert implement Buffer{} interface.
func (lb *LinearBuffer) Insert(dot int64, text []rune) (*LinearBuffer, error) {
	if text == nil {
		return lb, nil
	} else if lb == nil {
		return NewLinearBuffer(text), nil
	}

	left, right, err := lb.Split(dot)
	if err != nil {
		return nil, err
	}
	if left == nil {
		left = NewLinearBuffer([]rune(""))
	}
	if right == nil {
		right = NewLinearBuffer([]rune(""))
	}
	newlb := make([]rune, len(lb.Text)+len(text))
	copy(newlb, left.Text)
	copy(newlb[len(left.Text):], text)
	copy(newlb[len(left.Text)+len(text):], right.Text)
	return NewLinearBuffer(newlb), nil
}

// InsertIO implement Buffer{} interface.
func (lb *LinearBuffer) InsertIO(dot int64, text []rune) (*LinearBuffer, error) {
	return lb.Insert(dot, text)
}

// Delete implement Buffer{} interface.
func (lb *LinearBuffer) Delete(dot, n int64) (*LinearBuffer, error) {
	if lb == nil {
		return nil, v.ErrorBufferNil
	} else if dot < 0 || dot > int64(len(lb.Text)-1) {
		return nil, v.ErrorIndexOutofbound
	} else if end := dot + n; end < 0 || end > int64(len(lb.Text)) {
		return nil, v.ErrorIndexOutofbound
	}
	newt := make([]rune, int64(len(lb.Text))-n)
	copy(newt[:dot], lb.Text[:dot])
	copy(newt[dot:], lb.Text[dot+n:])
	newlb := NewLinearBuffer(newt)
	return newlb, nil
}

// DeleteIO implement Buffer{} interface.
func (lb *LinearBuffer) DeleteIO(dot, n int64) (*LinearBuffer, error) {
	return lb.Delete(dot, n)
}

// Stats implement Buffer{} interface.
func (lb *LinearBuffer) Stats() (stats v.Statistics, err error) {
	return
}
