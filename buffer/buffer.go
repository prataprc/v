package buffer

import "fmt"
import "errors"

var _ = fmt.Sprintf("dummy")

// MaxRuneWidth is sizeof(rune) datatype.
const MaxRuneWidth = 4

// Statistics is a key,value map of counters.
type Statistics map[string]interface{}

//-------------------
// Buffer error codes
//-------------------

// ErrorBufferNil says buffer is not initialized.
var ErrorBufferNil = errors.New("buffer.uninitialized")

// ErrorIndexOutofbound says access to buffer is outside its
// size.
var ErrorIndexOutofbound = errors.New("buffer.indexOutofbound")

// ErrorInvalidEncoding
var ErrorInvalidEncoding = errors.New("buffer.invalidEncoding")

// ErrorReadonlyBuffer says buffer cannot be changed.
var ErrorReadonlyBuffer = errors.New("editbuffer.ronly")

// ErrorOldestChange says there is not more change to undo.
var ErrorOldestChange = errors.New("editbuffer.oldestChange")

// ErrorLatestChange says there is no more change to redo.
var ErrorLatestChange = errors.New("editbuffer.latestChange")

// Buffer describes a buffer and APIs to access the buffer,
// where a buffer can be implemented as linear array, gap-buffer,
// rope-buffer, line-buffer etc.
//
// bCur - utf8 aligned cursor within the buffer starting from 0,
// where a value of N means there are N bytes before the
// cursor, 0 means start and len(buffer) means end.
type Buffer interface {

	//---- Rune APIs

	// Length return no. of runes in buffer.
	Length() int64

	// Slice return slice of runes of Bn bytes starting
	// from rune offset rCur.
	Slice(rCur, rn int64) Buffer

	// Runes return full content in buffer as rune-array.
	Runes() []rune

	// Concat adds another buffer element adjacent to the
	// current buffer.
	Concat(other Buffer) Buffer

	// Split this buffer at rCur, and return two equivalent buffer
	// elements, a rCur of value `n` would mean `n` runes to the
	// left buffer.
	Split(rCur int64) (left, right Buffer)

	// Insert addes one or more runes at rCur, where there would
	// be rCur runes to the left.
	// Returns a new buffer with inserted runes.
	Insert(rCur int64, text []rune) Buffer

	// Delete will remove `rn` runes after `rCur`.
	// Returns a new buffer with deleted runes.
	Delete(rCur int64, rn int64) Buffer

	// InsertIn addes, in place, one or more runes at rCur,
	// semantically pushing the runes at the rCur to the right.
	// Returns the same reference. NonPersistant API.
	InsertIn(rCur int64, text []rune) Buffer

	// DeleteIn deletes, in place, `rn` runes from the original
	// buffer after rCur.
	// Returns the same reference. NonPersistant API.
	DeleteIn(rCur int64, rn int64) Buffer

	//---- Search APIs

	// StreamFrom returns a RuneReader starting from `rCur`.
	StreamFrom(rCur int64) RuneReader

	// StreamCount returns a RuneReader starting from `rCur`
	// for `count` number of runes.
	StreamCount(rCur, count int64) RuneReader

	// BackStreamFrom returns a RuneReader starting from `rCur`,
	// streaming in the backward direction.
	BackStreamFrom(rCur int64) RuneReader

	// BackStreamCount returns a RuneReader starting from `rCur`,
	// in backward direction, for `count` number of runes.
	BackStreamCount(rCur, count int64) RuneReader

	//---- Byte APIs

	// Size return no. of bytes in buffer.
	Size() int64

	// Bytes returns full content for buffer as byte-array.
	Bytes() []byte

	// Stats return a key,value pair of interesting statistiscs.
	Stats() (Statistics, error)
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
