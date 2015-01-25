package buffer

import "errors"

// ErrorBufferNil says buffer is not initialized.
var ErrorBufferNil = errors.New("buffer.uninitialized")

// ErrorIndexOutofbound says access to buffer is outside its
// size.
var ErrorIndexOutofbound = errors.New("buffer.indexOutofbound")

// ErrorInvalidEncoding
var ErrorInvalidEncoding = errors.New("buffer.invalidEncoding")
