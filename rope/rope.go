package rope

import "fmt"
import "math"
import "github.com/prataprc/v"

var _ = fmt.Sprintf("dummy")

var RopeBufferCapacity = int64(64)

// RopeBuffer represents a persistent rope data structure.
type RopeBuffer struct {
	Dot    int64       `json:"dot,omitempty"`
	Text   []rune      `json:"text,omitempty"`
	Weight int64       `json:"weight,omitempty"`
	Len    int64       `json:"length,omitempty"`
	Cap    int64       `json:"capacity,omitempty"`
	Left   *RopeBuffer `json:"left,omitempty"`
	Right  *RopeBuffer `json:"right,omitempty"`
}

// NewRopebuffer returns a new buffer, initialized with text.
func NewRopebuffer(text []rune, capacity int64) *RopeBuffer {
	l := int64(len(text))
	txt := make([]rune, l)
	copy(txt, text)
	rb := &RopeBuffer{
		Dot:    0,
		Text:   txt,
		Weight: l,
		Len:    l,
		Cap:    capacity,
	}
	rb, _ = rb.build(capacity)
	return rb
}

// NewRopeLevel returns a new level of rope-buffer initialized
// with left and right.
func NewRopeLevel(
	weight, length, capacity int64, left, right *RopeBuffer) *RopeBuffer {
	return &RopeBuffer{
		Weight: weight, Len: length, Left: left, Right: right, Cap: capacity,
	}
}

// Cursor implement Buffer{} interface.
func (rb *RopeBuffer) Cursor() int64 {
	return rb.Dot
}

// Length implement Buffer{} interface.
func (rb *RopeBuffer) Length() (l int64, err error) {
	if rb == nil {
		return l, v.ErrorBufferNil
	}
	return rb.Len, nil
}

// Value implement Buffer{} interface.
func (rb *RopeBuffer) Value() []rune {
	acc := make([]rune, rb.Len)
	rb.report(0, rb.Len, acc)
	return acc
}

// Index implement Buffer{} interface.
func (rb *RopeBuffer) Index(dot int64) (ch rune, ok bool, err error) {
	if rb == nil {
		return ch, ok, v.ErrorBufferNil
	} else if dot < 0 || dot > rb.Len {
		return ch, ok, v.ErrorIndexOutofbound
	} else if dot == rb.Len {
		return ch, false, v.ErrorIndexOutofbound
	}

	if rb.isLeaf() {
		return rb.Text[dot], true, nil
	} else if dot >= rb.Weight {
		return rb.Right.Index(dot - rb.Weight)
	}
	return rb.Left.Index(dot)
}

// Substr implement Buffer{} interface.
func (rb *RopeBuffer) Substr(dot int64, n int64) (string, error) {
	if rb == nil {
		return "", nil
	} else if dot < 0 || dot > rb.Len {
		return "", v.ErrorIndexOutofbound
	} else if dot+n > rb.Len {
		return "", v.ErrorIndexOutofbound
	}

	acc := make([]rune, n)
	rb.report(dot, n, acc)
	return string(acc), nil
}

// Concat implement Buffer{} interface.
func (rb *RopeBuffer) Concat(right *RopeBuffer) (*RopeBuffer, error) {
	if rb == nil {
		return right, nil
	} else if right == nil {
		return rb, nil
	}
	return NewRopeLevel(rb.Len, rb.Len+right.Len, rb.Cap, rb, right), nil
}

// Split implement Buffer{} interface.
func (rb *RopeBuffer) Split(dot int64) (left, right *RopeBuffer, err error) {
	if rb == nil {
		return nil, nil, v.ErrorBufferNil

	} else if dot < 0 || dot > rb.Len {
		return nil, nil, v.ErrorIndexOutofbound
	}
	return rb.split(dot, nil)
}

// Insert implement Buffer{} interface.
func (rb *RopeBuffer) Insert(
	dot int64, text []rune, amend bool) (*RopeBuffer, error) {

	// TODO: Implement amend.
	if text == nil {
		return rb, nil

	} else if rb == nil {
		return NewRopebuffer(text, rb.Cap), nil
	}
	left, right, err := rb.Split(dot)
	if err != nil {
		return nil, err
	}
	x, err := left.Concat(NewRopebuffer(text, rb.Cap))
	if err != nil {
		return nil, err
	}
	return x.Concat(right)
}

// Delete implement Buffer{} interface.
func (rb *RopeBuffer) Delete(dot int64, n int64) (*RopeBuffer, error) {
	if rb == nil {
		return nil, v.ErrorBufferNil
	} else if l := rb.Len; dot < 0 || dot > int64(l-1) {
		return nil, v.ErrorIndexOutofbound
	} else if end := dot + n; end < 0 || end > int64(l) {
		return nil, v.ErrorIndexOutofbound
	}
	left, forRight, err := rb.Split(dot)
	if err != nil {
		return nil, err
	}
	_, right, err := forRight.Split(n)
	if err != nil {
		return nil, err
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

func (rb *RopeBuffer) build(capacity int64) (*RopeBuffer, error) {
	if rb.isLeaf() {
		if rb.Len > capacity {
			left, right, _ := rb.Split(rb.Len / 2)
			x, err := left.build(capacity)
			if err != nil {
				return nil, err
			}
			y, err := right.build(capacity)
			if err != nil {
				return nil, err
			}
			return x.Concat(y)
		}
	}
	return rb, nil
}

func (rb *RopeBuffer) split(
	dot int64, right *RopeBuffer) (*RopeBuffer, *RopeBuffer, error) {

	var err error

	if dot == rb.Weight { // exact
		if rb.isLeaf() {
			return rb, rb.Right, nil
		}
		return rb.Left, rb.Right, nil

	} else if dot > rb.Weight { // recurse on the right
		newRight, right, err := rb.Right.split(dot-rb.Weight, right)
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
		if dot > 0 {
			l := NewRopebuffer(rb.Text[0:dot], rb.Cap)
			r := NewRopebuffer(rb.Text[dot:len(rb.Text)], rb.Cap)
			return l, r, nil
		}
		r := NewRopebuffer(rb.Text[dot:len(rb.Text)], rb.Cap)
		return nil, r, nil
	}
	newLeft, right, err := rb.Left.split(dot, right)
	if err != nil {
		return nil, nil, err
	}
	right, err = right.Concat(rb.Right)
	if err != nil {
		return nil, nil, err
	}
	return newLeft, right, nil
}

func (rb *RopeBuffer) report(dot int64, n int64, acc []rune) {
	if dot > rb.Weight { // recurse to right
		rb.Right.report(dot-rb.Weight, n, acc)

	} else if rb.Weight >= dot+n { // the left branch has enough values
		if rb.isLeaf() {
			copy(acc, rb.Text[dot:dot+n])
		} else {
			rb.Left.report(dot, n, acc)
		}

	} else { // else split the work
		rb.Left.report(dot, rb.Weight-dot, acc[:rb.Weight-dot])
		rb.Right.report(0, dot+n-rb.Weight, acc[rb.Weight-dot:])
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
