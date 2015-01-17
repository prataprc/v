package v

import "fmt"
import "errors"

var _ = fmt.Sprintf("dummy")

// ErrorBufferNil says buffer is not initialized.
var ErrorBufferNil = errors.New("buffer.uninitialized")

// ErrorIndexOutofbound says access to buffer is outside its
// size.
var ErrorIndexOutofbound = errors.New("buffer.indexOutofbound")

type Statistics map[string]interface{}

// within a buffer all text is unicode and represented
// as rune.
// - characters can enter buffer, via cmd, only as rune.
// - characters exit buffer only as rune.
//
// dot - is the cursor within the buffer starting from 0,
// where a value of N means there are N runes before the
// cursor, 0 means start, max(int64) means end.
type Buffer interface {
	// Length return no. of runes in buffer.
	Length() (l int64, err error)

	// Value returns full content in buffer.
	Value() []rune

	// Index retrieves the rune at dot, where
	// dot starts from 0 till length of buffer.
	// if buffer size is 22
	// 0 - returns the first rune in buffer.
	// 21 - returns the last rune in buffer.
	// 22 - return ok as false.
	Index(dot int64) (ch rune, ok bool, err error)

	// Substr return a substring of length N after dot
	// number of runes from buffer.
	Substr(dot int64, N int64) (val []rune, err error)

	// Concat adds another buffer element adjacent to the
	// current buffer.
	Concat(other *Buffer) (buf Buffer, err error)

	// Split this buffer at dot, and return two equivalent buffer
	// elements, a dot of value N would mean N elements to the
	// left buffer.
	Split(dot int64) (left Buffer, right Buffer, err error)

	// Insert addes one or more runes at dot, semantically
	// pushing the runes at the dot to the right.
	// `amend` argument is a special case, where inserts
	// belonging to the same change can mutate the buffer.
	Insert(dot int64, text []rune, amend bool) (buf Buffer, err error)

	// Delete generates a new buffer by deleting N runes from
	// the original buffer after dot.
	Delete(dot int64, N int64) (buf Buffer, err error)

	// Stats return a key,value pair of interesting statistiscs.
	Stats() (stats Statistics, err error)
}

type Cmd struct {
	args []interface{}
	c    string
	zrgs []interface{}
}

// Commands
const (
	// insert mode.
	RuboutChar string = "k"
	RuboutWord        = "ctrl-w"
	RuboutLine        = "ctrl-u"
	// normal mode.
	DotForward    = "l" // {args{Int}, "l"}
	DotForwardTok = "w" // {args{Int}, "w"}
	DotLineUp     = "k" // {args{Int}, "k"}
	DotGoto       = "j" // {args{Int}, "j"}
	// ex-command.
	Ex = "exit"
)

// args and zrgs
const (
	BuStart string = "0" // beginning of buffer
	BuEnd   string = "z" // end of buffer
	BuAll   string = "h" // full buffer
)
