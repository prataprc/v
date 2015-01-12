package main

import "fmt"
import "errors"

var _ = fmt.Sprintf("dummy")

var errorBufferNil = errors.New("buffer.uninitialized")
var errorIndexOutofbound = errors.New("buffer.indexOutofbound")

// within a buffer all text is unicode and represented
// as rune.
// - characters can enter buffer, via cmd, only as rune.
// - characters exit buffer only as rune.

// dot - is the cursor within the buffer starting from 0,
// where a value of N means there are N runes before the
// cursor, 0 means start, max(int64) means end.

type Buffer interface{
    // Length return no. of runes in buffer.
    Length() int64

    // Value returns full content in buffer.
    Value() []rune

    // Index retrieves the rune at dot, where
    // dot starts from 1.
    Index(dot int64) rune

    // Concat adds another buffer element adjacent to the
    // current buffer.
    Concat(other *Buffer) *Buffer

    // Split this buffer at dot, and return two equivalent buffer
    // elements, a dot of value N would mean N elements to the
    // left buffer.
    Split(dot int64) (Buffer, Buffer)

    // Insert addes one or more runes at dot, semantically
    // pushing the runes at the dot to the right.
    // `amend` argument is a special case, where inserts
    // belonging to the same change can mutate the buffer.
    Insert(dot int64, text []rune, amend bool) Buffer

    // Delete generates a new buffer by deleting N runes from
    // the original buffer after dot.
    Delete(dot int64, N int64) Buffer

    // Substr return a substring of length N after dot
    // number of runes from buffer.
    Substr(dot int64, N int64) string
}
