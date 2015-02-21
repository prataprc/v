package buffer

import "fmt"
import "regexp"

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
	Runes() ([]rune, error)

	// RuneSlice return a substring of length `rn` runes after
	// bCur number of bytes from start. It also returns the
	// no.of bytes consume to decode `rn` runes.
	RuneSlice(bCur, rn int64) (runes []rune, size int64, err error)

	// Concat adds another buffer element adjacent to the
	// current buffer.
	Concat(other Buffer) (buf Buffer, err error)

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
	StreamFrom(bCur int64) RuneReader

	// StreamCount returns a RuneReader starting from `bCur`
	// for `count` number of runes.
	StreamCount(bCur, count int64) RuneReader

	// StreamTill returns a RuneReader starting from `bCur`
	// until `till` number of bytes decoded.
	StreamTill(bCur, till int64) RuneReader

	// BackStreamFrom returns a RuneReader starting from `bCur`,
	// streaming in the backward direction.
	BackStreamFrom(bCur int64) RuneReader

	// BackStreamCount returns a RuneReader starting from `bCur`,
	// in backward direction, for `count` number of runes.
	BackStreamCount(bCur, count int64) RuneReader

	// BackStreamTill returns a RuneReader starting from `bCur`,
	// in backward direction, until `till` number of bytes decoded.
	BackStreamTill(bCur, till int64) RuneReader

	// Stats return a key,value pair of interesting statistiscs.
	Stats() (stats Statistics, err error)
}

type RuneReader interface {
	// ReadRune will read one rune at a time from buffer,
	// until io.EOF is reached.
	ReadRune() (r rune, size int, err error)

	// Close() the reader.
	Close()
}

// iterator implements RuneReader interface{}.
type iterator func(finish bool) (r rune, size int, err error)

// ReadRune implements RuneReader interface{}.
func (fn iterator) ReadRune() (rune, int, error) {
	return fn(false)
}

// ReadRune implements RuneReader interface{}.
func (fn iterator) Close() {
	fn(true)
}

// Find `regex` pattern within buffer streamed by RuneReader
// streaming can be in any direction and regex is expected to be
// specified accordingly.
func Find(regex interface{}, r RuneReader) Finder {
	var re *regexp.Regexp
	var err error

	switch v := regex.(type) {
	case string:
		re, err = regexp.Compile(v)
		if err != nil {
			panic(err)
		}
	case *regexp.Regexp:
		re = v
	case func() *regexp.Regexp:
		re = v()
	default:
		panic("impossible situation")
	}

	return Finder(func() []int { return re.FindReaderIndex(r) })
}
