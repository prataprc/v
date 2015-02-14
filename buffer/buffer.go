package buffer

import "fmt"
import "io"

var _ = fmt.Sprintf("dummy")

const MaxRuneWidth = 10

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

	// RuneSlice return a substring of length `rn` runes after
	// bCur number of bytes from start. It also returns the
	// no.of bytes consume to decode `rn` runes.
	RuneSlice(bCur, rn int64) (runes []rune, size int64, err error)

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

	// Delete generates a new buffer by deleting `rn` runes from
	// the original buffer after bCur. Returns a new reference
	// without creating side-effects.
	Delete(bCur int64, rn int64) (buf Buffer, err error)

	// InsertIn addes, in place, one or more runes at bCur,
	// semantically pushing the runes at the bCur to the right.
	// Returns the same reference.
	// NonPersistant API.
	InsertIn(bCur int64, text []rune) (buf Buffer, err error)

	// DeleteIn deletes, in place, `rn` runes from the original
	// buffer after bCur. Returns the same reference.
	// NonPersistant API.
	DeleteIn(bCur int64, rn int64) (buf Buffer, err error)

	// StreamFrom returns a RuneReader starting from `bCur`.
	StreamFrom(bCur int64) io.RuneReader

	// StreamTill returns a RuneReader starting from `bCur` for
	// `count` number of runes.
	StreamTill(bCur, count int64) io.RuneReader

	// BackStreamFrom returns a RuneReader starting from `bCur`,
	// streaming in the backward direction.
	BackStreamFrom(bCur int64) io.RuneReader

	// BackStreamTill returns a RuneReader starting from `bCur` for
	// `count` number of runes, in the backward direction.
	BackStreamTill(bCur, count int64) io.RuneReader

	// Stats return a key,value pair of interesting statistiscs.
	Stats() (stats Statistics, err error)
}

// iterator implements io.RuneReader interface{}.
type iterator func() (r rune, size int, err error)

// ReadRune implements io.RuneReader interface{}.
func (fn iterator) ReadRune() (rune, int, error) {
	return fn()
}
