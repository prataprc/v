package rope

import "fmt"
import "github.com/prataprc/v"

var _ = fmt.Sprintf("dummy")

// LinearBuffer represents mutable sequence of runes as buffer.
type LinearBuffer []rune

// NewLinearBuffer returns a new buffer, initialized with text.
func NewLinearBuffer(text []rune) LinearBuffer {
	return LinearBuffer(text)
}

// Length implement Buffer{} interface.
func (lb LinearBuffer) Length() (n int64, err error) {
	if lb == nil {
		return n, v.ErrorBufferNil
	}
	return int64(len(lb)), nil
}

// Value implement Buffer{} interface.
func (lb LinearBuffer) Value() []rune {
	if lb == nil {
		return nil
	}
	return []rune(lb)
}

// Index implement Buffer{} interface.
func (lb LinearBuffer) Index(dot int64) (ch rune, ok bool, err error) {
	if lb == nil {
		return ch, false, v.ErrorBufferNil
	} else if dot < 0 || dot > int64(len(lb)) {
		return ch, false, v.ErrorIndexOutofbound
	} else if dot == int64(len(lb)) {
		return ch, false, nil
	}
	return lb[dot], true, nil
}

// Concat implement Buffer{} interface.
func (lb LinearBuffer) Concat(right LinearBuffer) (LinearBuffer, error) {
	if lb == nil {
		return right, nil
	} else if right == nil {
		return lb, nil
	}
	return append(lb, right...), nil
}

// Split implement Buffer{} interface.
func (lb LinearBuffer) Split(dot int64) (left, right LinearBuffer, err error) {
	if lb == nil {
		return left, right, v.ErrorBufferNil
	} else if dot < 0 || dot > int64(len(lb)) {
		return left, right, v.ErrorIndexOutofbound
	} else if dot == 0 {
		return nil, lb, nil
	} else if dot == int64(len(lb)) {
		return lb, nil, nil
	}
	left, right = make([]rune, dot), make([]rune, int64(len(lb))-dot)
	copy(left, lb[:dot])
	copy(right, lb[dot:])
	return left, right, nil
}

// Insert implement Buffer{} interface.
func (lb LinearBuffer) Insert(
	dot int64, text []rune, amend bool) (LinearBuffer, error) {

	if text == nil {
		return lb, nil

	} else if lb == nil {
		return NewLinearBuffer(text), nil
	}
	left, right, err := lb.Split(dot)
	if err != nil {
		return nil, err
	}
	newlb := make([]rune, len(lb)+len(text))
	copy(newlb, left)
	copy(newlb[len(left):], text)
	copy(newlb[len(left)+len(text):], right)
	return newlb, nil
}

// Delete implement Buffer{} interface.
func (lb LinearBuffer) Delete(dot int64, n int64) (LinearBuffer, error) {
	if lb == nil {
		return nil, v.ErrorBufferNil
	}
	l := int64(len(lb))
	copy(lb[dot:], lb[dot+n:])
	lb = lb[:l-n]
	return lb, nil
}

// Substr implement Buffer{} interface.
func (lb LinearBuffer) Substr(dot int64, n int64) (string, error) {
	if lb == nil {
		return "", nil
	}
	return string(lb[dot : dot+n]), nil
}

// Stats implement Buffer{} interface.
func (lb LinearBuffer) Stats() (stats v.Statistics, err error) {
	return
}
