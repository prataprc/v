package buffer

import "fmt"
import "io"
import "math"
import "unicode/utf8"

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
	newt := make([]byte, l)
	copy(newt, text)
	rb := &RopeBuffer{Text: newt, Weight: l, Len: l, Cap: capacity}
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

func (rb *RopeBuffer) String() string {
	return fmt.Sprintf("{W%d L%d C%d}", rb.Weight, rb.Len, rb.Cap)
}

//----------------------
// Buffer implementation
//----------------------

// Length implement Buffer{} interface.
func (rb *RopeBuffer) Length() (n int64, err error) {
	if rb == nil {
		return n, ErrorBufferNil
	}
	return rb.Len, nil
}

// Value implement Buffer{} interface.
func (rb *RopeBuffer) Value() []byte {
	if rb == nil {
		return nil
	} else if rb.Len == 0 {
		return []byte{}
	}
	acc := make([]byte, rb.Len)
	rb.value(0, rb.Len, acc)
	return acc
}

// Slice implement Buffer{} interface.
func (rb *RopeBuffer) Slice(bCur, bn int64) ([]byte, error) {
	if rb == nil {
		return nil, ErrorBufferNil
	} else if bn == 0 || rb.Len == 0 {
		return []byte{}, nil
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
func (rb *RopeBuffer) RuneAt(bCur int64) (ch rune, size int, err error) {
	if rb == nil {
		return ch, size, ErrorBufferNil
	} else if bCur < 0 || bCur >= rb.Len {
		return ch, size, ErrorIndexOutofbound
	}

	if rb.isLeaf() {
		ch, size := utf8.DecodeRune(rb.Text[bCur:])
		if ch == utf8.RuneError {
			return ch, size, ErrorInvalidEncoding
		}
		return ch, size, nil

	} else if bCur >= rb.Weight {
		return rb.Right.RuneAt(bCur - rb.Weight)
	}
	return rb.Left.RuneAt(bCur)
}

// Runes implement Buffer{} interface.
func (rb *RopeBuffer) Runes() ([]rune, error) {
	acc := make([]rune, rb.Len)
	count, size, err := rb.runes(0, rb.Len, acc)
	if err != nil {
		return nil, err
	} else if size != rb.Len {
		panic("mismatch in decoded bytes and length")
	}
	return acc[:count], nil
}

// RuneSlice implement Buffer{} interface.
func (rb *RopeBuffer) RuneSlice(bCur, rn int64) ([]rune, int64, error) {
	if rb == nil {
		return nil, 0, ErrorBufferNil
	} else if rn == 0 {
		return []rune{}, 0, nil
	} else if bCur < 0 || bCur >= rb.Len {
		return nil, 0, ErrorIndexOutofbound
	} else if rb.Len == 0 {
		return nil, 0, ErrorIndexOutofbound
	}

	var reader RuneReader
	if rn > 0 {
		reader = rb.StreamCount(bCur, rn)
	} else {
		reader = rb.BackStreamCount(bCur, rn)
	}
	defer reader.Close()

	var err error
	var r rune
	var sz int
	i, runes, size := 0, make([]rune, rn), int64(0)
	for {
		if r, sz, err = reader.ReadRune(); err == io.EOF {
			break
		} else if err == nil {
			runes[i] = r
			i, size = i+1, size+int64(sz)
		} else {
			return nil, 0, err
		}
	}
	return runes[:i], size, nil
}

// Concat implement Buffer{} interface.
func (rb *RopeBuffer) Concat(right Buffer) (Buffer, error) {
	rightbuf := right.(*RopeBuffer)
	if rb == nil {
		return rightbuf, nil
	} else if rightbuf == nil {
		return rb, nil
	}
	return NewRopeLevel(rb.Len+rightbuf.Len, rb, rightbuf), nil
}

// Split implement Buffer{} interface.
func (rb *RopeBuffer) Split(bCur int64) (left, right Buffer, err error) {
	if rb == nil {
		return nil, nil, ErrorBufferNil

	} else if bCur < 0 || bCur > rb.Len {
		return nil, nil, ErrorIndexOutofbound
	}
	return rb.split(bCur, nil)
}

// Insert implement Buffer{} interface.
func (rb *RopeBuffer) Insert(bCur int64, text []rune) (Buffer, error) {
	if text == nil {
		return rb, nil

	} else if rb == nil {
		return rb, ErrorBufferNil

	} else if bCur < 0 || bCur > rb.Len {
		return rb, ErrorIndexOutofbound
	}
	textb := []byte(string(text))
	if len(textb) == 0 { // nothing to insert
		return rb, nil
	}
	insrtRight, err := NewRopebuffer(textb, rb.Cap)
	if err != nil {
		return rb, err
	}

	left, right, err := rb.Split(bCur)
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
func (rb *RopeBuffer) Delete(bCur, rn int64) (Buffer, error) {
	if rb == nil {
		return rb, ErrorBufferNil
	} else if rn == 0 {
		return rb, nil
	} else if bCur < 0 || bCur > rb.Len-1 {
		return rb, ErrorIndexOutofbound
	}

	runes, size, err := rb.RuneSlice(bCur, rn)
	if err != nil {
		return rb, err
	} else if int64(len(runes)) != rn {
		return rb, ErrorIndexOutofbound
	} else if end := bCur + size; end < 0 || end > rb.Len {
		return rb, ErrorIndexOutofbound
	} else if bCur == 0 && size == rb.Len { // to delete entire buffer
		return NewRopebuffer([]byte{}, rb.Cap)
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

// InsertIn implement Buffer{} interface.
func (rb *RopeBuffer) InsertIn(bCur int64, text []rune) (Buffer, error) {
	textb := []byte(string(text)) // TODO: this could be inefficient
	if text == nil || len(textb) == 0 {
		return rb, nil
	} else if rb == nil {
		return rb, ErrorBufferNil
	} else if rb.Len == 0 {
		return NewRopebuffer(textb, rb.Cap)
	} else if bCur < 0 || bCur > rb.Len {
		return rb, ErrorIndexOutofbound
	}

	if rb.isLeaf() { // make inplace modification
		text := ioInsert(rb.Text, textb, bCur)
		newrb, err := NewRopebuffer(text, rb.Cap)
		if err != nil {
			return rb, err
		}
		rb.copyrefs(newrb)
		return rb, nil
	}
	if bCur >= rb.Weight {
		right, err := rb.Right.InsertIn(bCur-rb.Weight, text)
		if err != nil {
			return rb, err
		}
		rb.Right = right.(*RopeBuffer)
		rb.Len = rb.Left.Len + rb.Right.Len
		return rb, nil
	}
	left, err := rb.Left.InsertIn(bCur, text)
	if err != nil {
		return rb, err
	}
	rb.Left = left.(*RopeBuffer)
	rb.Weight, rb.Len = rb.Left.Len, rb.Left.Len+rb.Right.Len
	return rb, nil
}

// DeleteIn implement Buffer{} interface.
func (rb *RopeBuffer) DeleteIn(bCur, rn int64) (Buffer, error) {
	if rb == nil {
		return rb, ErrorBufferNil
	} else if rn == 0 {
		return rb, nil
	} else if bCur < 0 || bCur > (rb.Len-1) {
		return rb, ErrorIndexOutofbound
	}
	_, size, err := rb.RuneSlice(bCur, rn)
	if err != nil {
		return rb, err
	} else if end := bCur + size; end < 0 || end > rb.Len {
		return rb, ErrorIndexOutofbound
	}
	return rb.deleteIn(bCur, size)
}

// StreamFrom implement Buffer interface{}.
func (rb *RopeBuffer) StreamFrom(bCur int64) RuneReader {
	// clear up the corner cases.
	if bCur < 0 || rb.Len == 0 {
		return iterator(runeEnd)
	}
	return rb.runeIterator(bCur)
}

// StreamCount implement Buffer interface{}.
func (rb *RopeBuffer) StreamCount(bCur, count int64) RuneReader {
	// clear up the corner cases.
	if bCur < 0 || (rb.Len == 0) || (count <= 0) {
		return iterator(runeEnd)
	}
	return rb.runeIteratorCount(bCur, count)
}

// StreamTill implement Buffer interface{}.
func (rb *RopeBuffer) StreamTill(bCur, till int64) RuneReader {
	// clear up the corner cases.
	if bCur < 0 || (rb.Len == 0) || (till >= bCur) {
		return iterator(runeEnd)
	}
	return rb.runeIteratorTill(bCur, till)
}

// BackStreamFrom implement Buffer interface{}.
func (rb *RopeBuffer) BackStreamFrom(bCur int64) RuneReader {
	if bCur < 0 || rb.Len == 0 {
		return iterator(runeEnd)
	}
	return rb.runeRIterator(bCur)
}

// BackStreamCount implement Buffer interface{}.
func (rb *RopeBuffer) BackStreamCount(bCur, count int64) RuneReader {
	if bCur < 0 || (rb.Len == 0) || (count <= 0) {
		return iterator(runeEnd)
	}
	return rb.runeRIteratorCount(bCur, count)
}

// BackStreamTill implement Buffer interface{}.
func (rb *RopeBuffer) BackStreamTill(bCur, till int64) RuneReader {
	if bCur < 0 || (rb.Len == 0) || (till >= bCur) {
		return iterator(runeEnd)
	}
	return rb.runeRIteratorTill(bCur, till)
}

// Stats implement Buffer{} interface.
func (rb *RopeBuffer) Stats() (Statistics, error) {
	s := newRBStatistics()
	rb.stats(1, s)
	return Statistics(s), nil
}

// JohnnieWalker gets called for every leaf node in the
// rope-tree. If johnnie walks forward, bCur points to
// first/start rune in the buffer. If johnnie walks
// backward, bCur points to the last/start rune in the
// buffer.
type JohnnieWalker func(bCur int64, rb *RopeBuffer) bool

// Walk the rope-tree starting from `bCur`,
// walk until johnnie return false.
func (rb *RopeBuffer) Walk(bCur int64, walkFn JohnnieWalker) bool {
	if rb.isLeaf() && bCur < rb.Len {
		return walkFn(bCur, rb)

	} else if !rb.isLeaf() {
		if bCur >= rb.Weight && rb.Weight < rb.Len {
			return rb.Right.Walk(bCur-rb.Weight, walkFn)

		} else {
			var cont bool
			if rb.Left != nil {
				if cont = rb.Left.Walk(bCur, walkFn); cont && rb.Right != nil {
					cont = rb.Right.Walk(0, walkFn)
				}
			} else if rb.Right != nil {
				cont = rb.Right.Walk(0, walkFn)
			} else {
				return true
			}
			return cont
		}
	}
	return true
}

// WalkBack in the rope-tree starting from `bCur`.
// walk until johnnie return false.
func (rb *RopeBuffer) WalkBack(bCur int64, walkFn JohnnieWalker) bool {
	if rb.isLeaf() && bCur >= 0 {
		if bCur == rb.Weight {
			n, err := getRuneStart(rb.Text[:bCur], true /*reverse*/)
			if err == nil {
				return walkFn(n, rb)
			}
		} else {
			return walkFn(bCur, rb)
		}

	} else if !rb.isLeaf() {
		if bCur >= rb.Weight && rb.Weight < rb.Len {
			var cont bool
			if rb.Right != nil {
				cont = rb.Right.WalkBack(bCur-rb.Weight, walkFn)
				if cont && rb.Left != nil {
					cont = rb.Left.WalkBack(rb.Weight, walkFn)
				}
			} else if rb.Left != nil {
				cont = rb.Left.WalkBack(rb.Weight, walkFn)
			} else {
				return true
			}
			return cont

		} else if rb.Left != nil {
			return rb.Left.WalkBack(bCur, walkFn)
		}
	}
	return true
}

//----------------
// Local functions
//----------------

func (rb *RopeBuffer) clear() *RopeBuffer {
	rb.Text = []byte{}
	rb.Weight, rb.Len = int64(len(rb.Text)), int64(len(rb.Text))
	rb.Left, rb.Right = nil, nil
	return rb
}

func (rb *RopeBuffer) isLeaf() bool {
	return rb.Left == nil
}

func (rb *RopeBuffer) copyrefs(newrb *RopeBuffer) {
	rb.Text = newrb.Text
	rb.Weight, rb.Len, rb.Cap = newrb.Weight, newrb.Len, newrb.Cap
	rb.Left, rb.Right = newrb.Left, newrb.Right
}

func (rb *RopeBuffer) build(capacity int64) (*RopeBuffer, error) {
	var x, y *RopeBuffer

	if rb.isLeaf() && rb.Len > 0 && rb.Len > capacity {
		splitAt := rb.Len / 2
		n, err := getRuneStart(rb.Text[splitAt:], false /*reverse*/)
		if err != nil {
			return nil, err
		}
		splitAt = splitAt + n
		l, r, err := rb.Split(splitAt)
		if err != nil {
			return nil, err
		}
		left, right := l.(*RopeBuffer), r.(*RopeBuffer)
		if x, err = left.build(capacity); err != nil {
			return nil, err
		}
		if y, err = right.build(capacity); err != nil {
			return nil, err
		}
		buf, err := x.Concat(y)
		return buf.(*RopeBuffer), err
	}
	return rb, nil
}

func (rb *RopeBuffer) split(
	bCur int64, right *RopeBuffer) (*RopeBuffer, *RopeBuffer, error) {

	var err error
	if bCur == rb.Weight { // exact
		if rb.isLeaf() {
			return rb, rb.Right /*nil*/, nil
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
		return left.(*RopeBuffer), right, nil

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
	rbuf, err := right.Concat(rb.Right)
	if err != nil {
		return nil, nil, err
	}
	return newLeft, rbuf.(*RopeBuffer), nil
}

func (rb *RopeBuffer) value(bCur int64, n int64, acc []byte) {
	if rb != nil {
		if bCur >= rb.Weight { // recurse to right
			rb.Right.value(bCur-rb.Weight, n, acc)

		} else if rb.Weight >= bCur+n { // the left branch has enough values
			if rb.isLeaf() {
				copy(acc, rb.Text[bCur:bCur+n])
			} else if rb.Left != nil {
				rb.Left.value(bCur, n, acc)
			}

		} else { // else split the work
			leftN := rb.Weight - bCur
			if rb.Left != nil {
				rb.Left.value(bCur, leftN, acc[:leftN])
			}
			if rb.Right != nil {
				rb.Right.value(0, n-leftN, acc[leftN:])
			}
		}
	}
}

func (rb *RopeBuffer) runes(
	bCur int64, rn int64, acc []rune) (int64, int64, error) {

	if rb.isLeaf() {
		count, size, err := bytes2NRunes(rb.Text[bCur:], rn, acc)
		if err != nil {
			return 0, 0, err
		}
		return count, size, nil

	} else if bCur >= rb.Weight { // recurse to right
		count, size, err := rb.Right.runes(bCur-rb.Weight, rn, acc[:rn])
		return count, size, err
	}
	// else split the work
	lcount, lsize, err := rb.Left.runes(bCur, rn, acc)
	if err != nil {
		return 0, 0, err

	} else if lcount < rn {
		if bCur+lsize != rb.Weight {
			err := fmt.Errorf(
				"mismatch in boundary %v, but %v", rb.Weight, bCur+lsize)
			panic(err)
		}
		rcount, rsize, err := rb.Right.runes(0, rn-lcount, acc[lcount:rn])
		return lcount + rcount, lsize + rsize, err
	}
	return lcount, lsize, err
}

func (rb *RopeBuffer) deleteIn(bCur, bn int64) (*RopeBuffer, error) {
	if rb.isLeaf() {
		rb.Text = ioDelete(rb.Text, bCur, bn)
		rb.Weight, rb.Len = int64(len(rb.Text)), int64(len(rb.Text))
		return rb, nil
	}

	var err error
	var left, right *RopeBuffer

	if bCur >= rb.Weight { // go right
		right, err = rb.Right.deleteIn(bCur-rb.Weight, bn)
		if err != nil {
			return rb, err
		}
		rb.Right, rb.Len = right, rb.Len-bn
		return rb, nil

	} else if bCur+bn < rb.Weight { // delete affects only the left path.
		left, err = rb.Left.deleteIn(bCur, bn)
		if err != nil {
			return rb, err
		}
		rb.Left, rb.Len, rb.Weight = left, rb.Len-bn, rb.Weight-bn
		return rb, nil
	}

	leftbn, rightbn := rb.Weight-bCur, bCur+bn-rb.Weight
	if rb.Left != nil {
		if left, err = rb.Left.deleteIn(bCur, leftbn); err != nil {
			return rb, err
		}
	}
	if rb.Right != nil {
		if right, err = rb.Right.deleteIn(0, rightbn); err != nil {
			return rb, err
		}
	}
	rb.Len, rb.Weight = rb.Len-bn, rb.Weight-leftbn
	rb.Left, rb.Right = left, right
	return rb, nil
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

func walkRoutine(
	rb *RopeBuffer, bCur int64, ch chan []interface{}, killch chan bool) {

	rb.Walk(bCur, func(bCur int64, leaf *RopeBuffer) bool {
		select {
		case ch <- []interface{}{bCur, leaf}:
			return true
		case <-killch:
			return false
		}
	})
	close(ch)
}

func walkBackRoutine(
	rb *RopeBuffer, bCur int64, ch chan []interface{}, killch chan bool) {

	rb.WalkBack(bCur, func(bCur int64, leaf *RopeBuffer) bool {
		select {
		case ch <- []interface{}{bCur, leaf}:
			return true
		case <-killch:
			return false
		}
	})
	close(ch)
}

func walkLeaf(ch chan []interface{}) func() (int64, *RopeBuffer) {
	return func() (int64, *RopeBuffer) {
		if iterVal, ok := <-ch; ok {
			off, leaf := iterVal[0].(int64), iterVal[1].(*RopeBuffer)
			return off, leaf
		}
		return 0, nil
	}
}

func runeEnd(finish bool) (r rune, size int, err error) {
	return r, size, io.EOF
}

// iterate on runes in buffer starting from `bCur`.
func (rb *RopeBuffer) runeIterator(bCur int64) iterator {
	ch, killch := make(chan []interface{}), make(chan bool)
	go walkRoutine(rb, bCur, ch, killch)
	nextLeaf := walkLeaf(ch)
	off, leaf := nextLeaf()
	return iterator(func(finish bool) (r rune, size int, err error) {
		if (leaf == nil) || finish {
			defer func() { recover() }()
			close(killch)
			return r, size, io.EOF

		} else if off < leaf.Len {
			r, size = utf8.DecodeRune(leaf.Text[off:])
			off += int64(size)
			if off == leaf.Len {
				off, leaf = nextLeaf()
			} else if off > leaf.Len {
				panic("impossible situation")
			}
			return r, size, nil
		}
		panic("impossible situation")
	})
}

// iterate on count runes in buffer starting from `bCur`.
func (rb *RopeBuffer) runeIteratorCount(bCur, count int64) iterator {
	ch, killch := make(chan []interface{}), make(chan bool)
	go walkRoutine(rb, bCur, ch, killch)
	nextLeaf := walkLeaf(ch)
	off, leaf := nextLeaf()
	return iterator(func(finish bool) (r rune, size int, err error) {
		if (leaf == nil) || (count <= 0) || finish {
			defer func() { recover() }()
			close(killch)
			return r, size, io.EOF

		} else if off < leaf.Len {
			r, size = utf8.DecodeRune(leaf.Text[off:])
			off += int64(size)
			if off == leaf.Len {
				off, leaf = nextLeaf()
			} else if off > leaf.Len {
				panic("impossible situation")
			}
			count--
			return r, size, nil
		}
		panic("impossible situation")
	})
}

// iterate `till` offset in buffer starting from `bCur`.
func (rb *RopeBuffer) runeIteratorTill(bCur, till int64) iterator {
	ch, killch := make(chan []interface{}), make(chan bool)
	go walkRoutine(rb, bCur, ch, killch)
	nextLeaf := walkLeaf(ch)
	off, leaf := nextLeaf()
	return iterator(func(finish bool) (r rune, size int, err error) {
		if (leaf == nil) || (bCur >= till) || finish {
			defer func() { recover() }()
			close(killch)
			return r, size, io.EOF

		} else if off < leaf.Len {
			r, size = utf8.DecodeRune(leaf.Text[off:])
			off += int64(size)
			till += int64(size)
			if off == leaf.Len {
				off, leaf = nextLeaf()
			} else if off > leaf.Len {
				panic("impossible situation")
			}
			return r, size, nil
		}
		panic("impossible situation")
	})
}

// iterate on runes in buffer starting from `bCur`,
// in reverse direction.
func (rb *RopeBuffer) runeRIterator(bCur int64) iterator {
	ch, killch := make(chan []interface{}), make(chan bool)
	go walkBackRoutine(rb, bCur, ch, killch)
	prevLeaf := walkLeaf(ch)
	off, leaf := prevLeaf()
	return iterator(func(finish bool) (r rune, size int, err error) {
		if leaf == nil || finish {
			defer func() { recover() }()
			close(killch)
			return r, size, io.EOF
		}
		r, size = utf8.DecodeRune(leaf.Text[off:])
		if off <= 0 {
			off, leaf = prevLeaf()
			return r, size, nil

		} else if off, err = getRuneStart(leaf.Text[:off], true); err == nil {
			return r, size, nil
		}
		return r, size, err
	})
}

// iterate on count runes in buffer starting from `bCur`,
// in reverse direction.
func (rb *RopeBuffer) runeRIteratorCount(bCur, count int64) iterator {
	ch, killch := make(chan []interface{}), make(chan bool)
	go walkBackRoutine(rb, bCur, ch, killch)
	prevLeaf := walkLeaf(ch)
	off, leaf := prevLeaf()
	return iterator(func(finish bool) (r rune, size int, err error) {
		if (leaf == nil) || (count == 0) || finish {
			defer func() { recover() }()
			close(killch)
			return r, size, io.EOF
		}
		r, size = utf8.DecodeRune(leaf.Text[off:])
		count--
		if off <= 0 {
			off, leaf = prevLeaf()
			return r, size, nil

		} else if off, err = getRuneStart(leaf.Text[:off], true); err == nil {
			return r, size, nil
		}
		return r, size, err
	})
}

// iterate `till` offset in buffer starting from `bCur`,
// in reverse direction.
func (rb *RopeBuffer) runeRIteratorTill(bCur, till int64) iterator {
	ch, killch := make(chan []interface{}), make(chan bool)
	go walkBackRoutine(rb, bCur, ch, killch)
	prevLeaf := walkLeaf(ch)
	off, leaf := prevLeaf()
	return iterator(func(finish bool) (r rune, size int, err error) {
		if (leaf == nil) || (bCur == 0) || (bCur < till) || finish {
			defer func() { recover() }()
			close(killch)
			return r, size, io.EOF
		}
		r, size = utf8.DecodeRune(leaf.Text[off:])
		bCur -= int64(size)
		if off <= 0 {
			off, leaf = prevLeaf()
			return r, size, nil
		} else if off, err = getRuneStart(leaf.Text[:off], true); err == nil {
			return r, size, nil
		}
		return r, size, err
	})
}

func (rb *RopeBuffer) printTree(prefix string) {
	if rb.isLeaf() {
		fmt.Printf("%sleaf %s\n", prefix, rb.String())
	} else {
		fmt.Printf("%s%s\n", prefix, rb.String())
		if rb.Left != nil {
			rb.Left.printTree(prefix + "  ")
		}
		if rb.Right != nil {
			rb.Right.printTree(prefix + "  ")
		}
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

func ioInsert(dest, text []byte, bCur int64) []byte {
	leftSl := make([]byte, int64(len(dest))-bCur)
	copy(leftSl, dest[bCur:])
	dest = append(dest[:bCur], text...)
	dest = append(dest, leftSl...)
	return dest
}

func ioDelete(dest []byte, bCur int64, n int64) []byte {
	l := int64(len(dest))
	copy(dest[bCur:], dest[bCur+n:])
	return dest[:l-n]
}
