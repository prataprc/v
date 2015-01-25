package buffer

import "fmt"
import "math"
import "unicode/utf8"

var _ = fmt.Sprintf("dummy")

var RopeBufferCapacity = int64(64)

// RopeBuffer represents a persistent rope data structure.
type RopeBuffer struct { // buffer implementation.
	Text   []byte      `json:"text,omitempty"`
	Weight int64       `json:"weight,omitempty"`
	Len    int64       `json:"length,omitempty"`
	Cap    int64       `json:"capacity,omitempty"`
	Left   *RopeBuffer `json:"left,omitempty"`
	Right  *RopeBuffer `json:"right,omitempty"`
}

// NewRopebuffer returns a new buffer, initialized with text.
// Capacity will decide the maximum length the buffer can hold
// beyond which it splits. The buffer always splits at unicode
// boundary.
func NewRopebuffer(text []byte, capacity int64) (*RopeBuffer, error) {
	l := int64(len(text))
	rb := &RopeBuffer{Text: text, Weight: l, Len: l, Cap: capacity}
	return rb.build(capacity)
}

// NewRopeLevel returns a new level of rope-buffer initialized
// with left and right. Uses the length of left-buffer as weight
// and left-buffer capacity as its own capacity.
func NewRopeLevel(length int64, left, right *RopeBuffer) *RopeBuffer {
	return &RopeBuffer{
		Weight: left.Len, Len: length, Left: left, Right: right, Cap: left.Cap,
	}
}

//----------------------
// Buffer implementation
//----------------------

// Length implement Buffer{} interface.
func (rb *RopeBuffer) Length() (l int64, err error) {
	if rb == nil {
		return l, ErrorBufferNil
	}
	return rb.Len, nil
}

// Value implement Buffer{} interface.
func (rb *RopeBuffer) Value() []byte {
	acc := make([]byte, rb.Len)
	rb.value(0, rb.Len, acc)
	return acc
}

// Slice implement Buffer{} interface.
func (rb *RopeBuffer) Slice(bCur, bn int64) ([]byte, error) {
	if rb == nil {
		return nil, nil
	} else if bCur < 0 || bCur > rb.Len {
		return nil, ErrorIndexOutofbound
	} else if bCur+bn > rb.Len {
		return nil, ErrorIndexOutofbound
	}

	acc := make([]byte, bn)
	rb.value(bCur, bn, acc)
	return acc, nil
}

// RuneAt implement Buffer{} interface.
func (rb *RopeBuffer) RuneAt(bCur int64) (ch rune, size int64, err error) {
	if rb == nil {
		return ch, size, ErrorBufferNil
	} else if bCur < 0 || bCur >= rb.Len {
		return ch, size, ErrorIndexOutofbound
	}

	if rb.isLeaf() {
		ch, size := utf8.DecodeRune(rb.Text[bCur:])
		if ch == utf8.RuneError {
			return ch, int64(size), ErrorInvalidEncoding
		}
		return ch, int64(size), nil

	} else if bCur >= rb.Weight {
		return rb.Right.RuneAt(bCur - rb.Weight)
	}
	return rb.Left.RuneAt(bCur)
}

// Runes implement Buffer{} interface.
func (rb *RopeBuffer) Runes() ([]rune, error) {
	acc := make([]rune, 0, rb.Len/8)
	_, size, err := rb.runes(0, rb.Len, acc)
	if err != nil {
		return nil, err
	} else if size != rb.Len {
		panic("mismatch in decoded bytes and length")
	}
	return acc, nil
}

// RuneSlice implement Buffer{} interface.
func (rb *RopeBuffer) RuneSlice(bCur, rn int64) ([]rune, int64, error) {
	if rb == nil {
		return nil, 0, ErrorBufferNil
	} else if bCur < 0 || bCur > rb.Len {
		return nil, 0, ErrorIndexOutofbound
	}

	acc := make([]rune, 0, rn)
	count, size, err := rb.runes(bCur, rn, acc)
	if err != nil {
		return nil, 0, err
	} else if count < rn && bCur+size != rb.Len {
		panic("mismatch in decoded bytes and length")
	}
	return acc, size, nil
}

// Concat implement Buffer{} interface.
func (rb *RopeBuffer) Concat(right *RopeBuffer) (*RopeBuffer, error) {
	if rb == nil {
		return right, nil
	} else if right == nil {
		return rb, nil
	}
	return NewRopeLevel(rb.Len+right.Len, rb, right), nil
}

// Split implement Buffer{} interface.
func (rb *RopeBuffer) Split(bCur int64) (left, right *RopeBuffer, err error) {
	if rb == nil {
		return nil, nil, ErrorBufferNil

	} else if bCur < 0 || bCur > rb.Len {
		return nil, nil, ErrorIndexOutofbound
	}
	return rb.split(bCur, nil)
}

// Insert implement Buffer{} interface.
func (rb *RopeBuffer) Insert(bCur int64, text []rune) (*RopeBuffer, error) {
	if text == nil {
		return rb, nil

	} else if rb == nil {
		return NewRopebuffer([]byte(string(text)), rb.Cap)

	} else if bCur < 0 || bCur > rb.Len {
		return rb, ErrorIndexOutofbound
	}
	left, right, err := rb.Split(bCur)
	if err != nil {
		return rb, err
	}
	insrtRight, err := NewRopebuffer([]byte(string(text)), rb.Cap)
	if err != nil {
		return rb, err
	}
	x, err := left.Concat(insrtRight)
	if err != nil {
		return rb, err
	}
	return x.Concat(right)
}

// Delete implement Buffer{} interface.
func (rb *RopeBuffer) Delete(bCur, rn int64) (*RopeBuffer, error) {
	if rb == nil {
		return rb, ErrorBufferNil

	} else if bCur < 0 || bCur > int64(rb.Len-1) {
		return rb, ErrorIndexOutofbound

	}

	_, size, err := rb.RuneSlice(bCur, rn)
	if err != nil {
		return rb, err
	} else if end := bCur + size; end < 0 || end > int64(rb.Len) {
		return rb, ErrorIndexOutofbound
	}

	left, forRight, err := rb.Split(bCur)
	if err != nil {
		return rb, err
	}
	_, right, err := forRight.Split(size)
	if err != nil {
		return rb, err
	}
	return left.Concat(right)
}

// Stats implement Buffer{} interface.
func (rb *RopeBuffer) Stats() rbStats {
	s := newRBStatistics()
	rb.stats(1, s)
	return s
}

//----------------
// Local functions
//----------------

func (rb *RopeBuffer) isLeaf() bool {
	return rb.Left == nil
}

func (rb *RopeBuffer) io(src, text []rune, dot int64) []rune {
	l := int64(len(text))
	newtext := make([]rune, len(text)+len(src))
	copy(newtext[:dot], src[:dot])
	copy(newtext[dot:l], text)
	copy(newtext[dot+l:], src[dot:])
	return newtext
}

func (rb *RopeBuffer) build(capacity int64) (*RopeBuffer, error) {
	var left, right, x, y *RopeBuffer

	if rb.isLeaf() && rb.Len > 0 && rb.Len > capacity {
		n, err := getRuneStart(rb.Text[capacity:])
		if err != nil {
			return nil, err
		}
		splitAt := rb.Len + n
		if left, right, err = rb.Split(splitAt / 2); err != nil {
			return nil, err
		}
		if x, err = left.build(capacity); err != nil {
			return nil, err
		}
		if y, err = right.build(capacity); err != nil {
			return nil, err
		}
		return x.Concat(y)
	}
	return rb, nil
}

func (rb *RopeBuffer) split(bCur int64, right *RopeBuffer) (*RopeBuffer, *RopeBuffer, error) {
	var err error

	if bCur == rb.Weight { // exact
		if rb.isLeaf() {
			return rb, rb.Right, nil
		}
		return rb.Left, rb.Right, nil

	} else if bCur > rb.Weight { // recurse on the right
		newRight, right, err := rb.Right.split(bCur-rb.Weight, right)
		if err != nil {
			return nil, nil, err
		}
		left, err := rb.Left.Concat(newRight)
		if err != nil {
			return nil, nil, err
		}
		return left, right, nil

	}
	// recurse on the left
	if rb.isLeaf() { // splitting leaf at index
		if bCur > 0 {
			l, err := NewRopebuffer(rb.Text[0:bCur], rb.Cap)
			if err != nil {
				return nil, nil, err
			}
			r, err := NewRopebuffer(rb.Text[bCur:len(rb.Text)], rb.Cap)
			if err != nil {
				return nil, nil, err
			}
			return l, r, nil
		}
		r, err := NewRopebuffer(rb.Text[bCur:len(rb.Text)], rb.Cap)
		if err != nil {
			return nil, nil, err
		}
		return nil, r, nil
	}
	newLeft, right, err := rb.Left.split(bCur, right)
	if err != nil {
		return nil, nil, err
	}
	right, err = right.Concat(rb.Right)
	if err != nil {
		return nil, nil, err
	}
	return newLeft, right, nil
}

func (rb *RopeBuffer) value(bCur int64, n int64, acc []byte) {
	if bCur > rb.Weight { // recurse to right
		rb.Right.value(bCur-rb.Weight, n, acc)

	} else if rb.Weight >= bCur+n { // the left branch has enough values
		if rb.isLeaf() {
			copy(acc, rb.Text[bCur:bCur+n])
		} else {
			rb.Left.value(bCur, n, acc)
		}

	} else { // else split the work
		leftN := rb.Weight - bCur
		rb.Left.value(bCur, leftN, acc[:leftN])
		rb.Right.value(0, n-leftN, acc[leftN:])
	}
}

func (rb *RopeBuffer) runes(bCur int64, rn int64, acc []rune) (int64, int64, error) {
	if rb.isLeaf() {
		count, size, err := bytes2RunesN(rb.Text[bCur:], rn, acc)
		if err != nil {
			return 0, 0, err
		}
		return count, size, nil

	} else if bCur > rb.Weight { // recurse to right
		return rb.Right.runes(bCur-rb.Weight, rn, acc)

	} else { // else split the work
		lcount, lsize, err := rb.Left.runes(bCur, rn, acc)
		if err != nil {
			return 0, 0, err
		} else if bCur+lsize != rb.Weight {
			panic("mismatch in boundary")
		}
		rcount, rsize, err := rb.Right.runes(bCur+lsize, rn-lcount, acc[lcount:])
		return lcount + rcount, lsize + rsize, err
	}
}

func (rb *RopeBuffer) stats(depth int64, s rbStats) {
	if rb.isLeaf() {
		s.incLeaves()
		s.statLevel(depth)
		s.statSizeCap(len(rb.Text), cap(rb.Text))

	} else {
		s.incNodes()
		rb.Left.stats(depth+1, s)
		rb.Right.stats(depth+1, s)
	}
}

//--------------------
// statistics template
//--------------------

type rbStats map[string]interface{}

func newRBStatistics() rbStats {
	return rbStats{
		"leafs":        int64(0),     // no. of leaf nodes
		"nodes":        int64(0),     // no. of intermediate nodes
		"length":       int64(0),     // length of useful content in buffer
		"capacity":     int64(0),     // length of useful content in buffer
		"minLevel":     int64(0),     // minimum level in the tree
		"maxLevel":     int64(0),     // maximum level in the tree
		"meanLevel":    float64(0.0), // mean level in the tree
		"deviantLevel": float64(0.0), // deviation in tree depth level
	}
}

func (s rbStats) statSizeCap(size, capacity int) {
	s["length"] = s["length"].(int64) + int64(size)
	s["capacity"] = s["capacity"].(int64) + int64(capacity)
}

func (s rbStats) incLength(size int64) {
	s["length"] = s["length"].(int64) + size
}

func (s rbStats) statLevel(depth int64) {
	s["minLevel"] = int64(1)
	if l := s["maxLevel"].(int64); l < depth {
		s["maxLevel"] = depth
	}
	avg, n := s["meanLevel"].(float64), s["leafs"].(int64)
	avgn := s.incAvg(avg, n, depth)
	varc := s["deviantLevel"].(float64)
	varcn := s.incVariance(avg, varc, n, depth)
	s["meanLevel"], s["deviantLevel"] = avgn, varcn
}

func (s rbStats) incLeaves() {
	s["leafs"] = s["leafs"].(int64) + 1
}

func (s rbStats) incNodes() {
	s["nodes"] = s["nodes"].(int64) + 1
}

// compute incremental average for every new element added
// to the sample set.
func (s rbStats) incAvg(avg float64, n, an int64) float64 {
	return (float64(n-1)*avg + float64(an)) / float64(n)
}

// compute incremental variance for every new element added
// to the sample set.
func (s rbStats) incVariance(avg, varc float64, n, an int64) float64 {
	avgn := s.incAvg(avg, n, an)
	avgdiff := (avg - avgn) * (avg - avgn)
	varndiff := (float64(an) - avgn) * (float64(an) - avgn)
	varcn := float64(n-2)*(varc*varc) + float64(n-1)*avgdiff + varndiff
	return math.Sqrt(varcn)
}
