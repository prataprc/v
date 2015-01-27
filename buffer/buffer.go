package buffer

import "fmt"

var _ = fmt.Sprintf("dummy")

type Statistics map[string]interface{}

// Buffer describes a buffer and access into buffer, where a
// buffer can be implemented as linear array, gap-buffer,
// rope-buffer etc.
//
// Within a buffer text is maintained as byte-array.
//
// bCur - unicode aligned cursor within the buffer starting from 0,
// where a value of N means there are N bytes before the
// cursor, 0 means start and z len(buffer) means.
type Buffer interface {
	// Length return no. of bytes in buffer.
	Length() (l int64, err error)

	// Value returns full content in buffer as byte-array.
	Value() []byte

	// Slice return a substring of length Bn bytes after
	// bCur number of bytes from start.
	Slice(bCur, Bn int64) (val []byte, err error)

	// RuneAt retrieves the rune at bCur and no.of bytes to
	// decode rune, where bCur starts from 0 till length of
	// buffer. If buffer size is 22
	// 0  - returns the first rune in buffer.
	// 21 - returns the last rune in buffer.
	// 22 - return ErrorIndexOutofbound
	RuneAt(bCur int64) (ch rune, size int, err error)

	// Runes return full content in buffer as rune-array.
	Runes() ([]byte, error)

	// RuneSlice return a substring of length Rn runes after
	// bCur number of bytes from start. It also returns the
	// no.of bytes consume to decode Rn runes.
	RuneSlice(bCur, Rn int64) (runes []rune, size int64, err error)

	// Concat adds another buffer element adjacent to the
	// current buffer.
	Concat(other *Buffer) (buf Buffer, err error)

	// Split this buffer at bCur, and return two equivalent buffer
	// elements, a bCur of value Bn would mean Bn bytes to the
	// left buffer.
	Split(bCur int64) (left Buffer, right Buffer, err error)

	// Insert addes one or more runes at bCur, semantically
	// pushing the runes at the bCur to the right. Returns a
	// new reference without creating side-effects.
	Insert(bCur int64, text []rune) (buf Buffer, err error)

	// Delete generates a new buffer by deleting Rn runes from
	// the original buffer after bCur. Returns a new reference
	// without creating side-effects.
	Delete(bCur int64, Rn int64) (buf Buffer, err error)

	// StreamFrom returns a RuneReader starting from `bCur`.
	//StreamFrom(bCur int64) io.RuneReader

	// Stats return a key,value pair of interesting statistiscs.
	Stats() (stats Statistics, err error)
}

// iterator implements io.RuneReader interface{}.
type iterator func() (r rune, size int, err error)

// ReadRune implements io.RuneReader interface{}.
func (fn iterator) ReadRune() (rune, int, error) {
	return fn()
}
