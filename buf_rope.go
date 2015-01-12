package main

import "fmt"

var _ = fmt.Sprintf("dummy")

var RopeBufferCapacity = int64(64)

// RopeBuffer represents a persistent rope data structure.
type RopeBuffer struct {
    Text   []rune       `json:"text,omitempty"`
    Weight int64        `json:"weight,omitempty"`
    Len    int64        `json:"length,omitempty"`
    Left   *RopeBuffer  `json:"left,omitempty"`
    Right  *RopeBuffer  `json:"right,omitempty"`
}

// NewRopebuffer returns a new buffer, initialized with text.
func NewRopebuffer(text []rune, capacity int64) *RopeBuffer {
    l := int64(len(text))
    txt := make([]rune, l)
    copy(txt, text)
    b := &RopeBuffer{Text: txt, Weight: l, Len: l}
    return b.build(capacity)
}

// NewRopeLevel returns a new level of rope-buffer initialized
// with left and right.
func NewRopeLevel(weight, length int64, left, right *RopeBuffer) *RopeBuffer {
    return &RopeBuffer{Weight: weight, Len: length, Left: left, Right: right}
}

// Length implements Buffer{} interface.
func (b *RopeBuffer) Length() int64 {
    if b == nil {
        panic(errorBufferNil)
    }
    return b.Len
}

// Value implements Buffer{} interface.
func (b *RopeBuffer) Value() []rune {
    acc := make([]rune, b.Len)
    b.report(0, b.Len, acc)
    return acc
}

// Index implements Buffer{} interface.
func (b *RopeBuffer) Index(dot int64) rune {
    if b == nil {
        panic(errorBufferNil)
    } else if dot < 1 || dot > b.Len {
        panic(errorIndexOutofbound)
    }

    if b.isLeaf() {
        return b.Text[dot-1]
    } else if dot > b.Weight {
        return b.Right.Index(dot - b.Weight)
    } else {
        return b.Left.Index(dot)
    }
}

// Concat implements Buffer{} interface.
func (b *RopeBuffer) Concat(right *RopeBuffer) *RopeBuffer {
    if b == nil {
        return right
    } else if right == nil {
        return b
    }
    return NewRopeLevel(b.Len, b.Len+right.Len, b, right)
}

// Split implements Buffer{} interface.
func (b *RopeBuffer) Split(dot int64) (left, right *RopeBuffer) {
    if b == nil {
        panic(errorBufferNil)

    } else if dot < 0 || dot > b.Len {
        panic(errorIndexOutofbound)
    }
    return b.split(dot, right)
}

// Insert implements Buffer{} interface.
func (b *RopeBuffer) Insert(dot int64, text []rune, amend bool) *RopeBuffer {
    // TODO: Implement amend.
    if text == nil {
        return b

    } else if b == nil {
        return NewRopebuffer(text, RopeBufferCapacity)
    }
    left, right := b.Split(dot)
    return left.Concat(NewRopebuffer(text, RopeBufferCapacity)).Concat(right)
}

// Delete implements Buffer{} interface.
func (b *RopeBuffer) Delete(dot int64, n int64) *RopeBuffer {
    if b == nil {
        panic(errorBufferNil)
    }
    left, forRight := b.Split(dot)
    _, right := forRight.Split(n)
    return left.Concat(right)
}

func (b *RopeBuffer) Substr(dot int64, n int64) string {
    if b == nil {
        return ""
    }
    acc := make([]rune, n)
    b.report(dot, n, acc)
    return string(acc)
}

//----------------
// Local functions
//----------------

func (b *RopeBuffer) isLeaf() bool {
    return b.Left == nil
}

func (b *RopeBuffer) build(capacity int64) *RopeBuffer {
    if b.isLeaf() {
        if b.Len > capacity {
            left, right := b.Split(b.Len/2)
            return left.build(capacity).Concat(right.build(capacity))
        }
    }
    return b
}

func (b *RopeBuffer) split(dot int64, right *RopeBuffer) (l, r *RopeBuffer) {
    if dot == b.Weight { // exact
        if b.isLeaf() {
            return b, b.Right
        }
        return b.Left, b.Right

    } else if dot > b.Weight { // recurse on the right
        newRight, right := b.Right.split(dot-b.Weight, right)
        return b.Left.Concat(newRight), right

    } else { // recurse on the left
        if b.isLeaf() { // splitting leaf at index
            if dot > 0 {
                l = NewRopebuffer(b.Text[0:dot], RopeBufferCapacity)
            }
            r = NewRopebuffer(b.Text[dot:len(b.Text)], RopeBufferCapacity)
            return l, r

        } else {
            newLeft, right := b.Left.split(dot, right)
            return newLeft, right.Concat(b.Right)
        }
    }
}

func (b *RopeBuffer) report(dot int64, n int64, acc []rune) {
    if dot > b.Weight { // recurse to right
        b.Right.report(dot-b.Weight, n, acc)
    } else if b.Weight >= dot+n { // the left branch has enough values
        if b.isLeaf() {
            copy(acc, b.Text[dot:dot+n])
        } else {
            b.Left.report(dot, n, acc)
        }

    } else { // else split the work
        b.Left.report(dot, b.Weight-dot, acc[:b.Weight])
        b.Right.report(0, dot+n-b.Weight, acc[b.Weight:])
    }
}
