package buffer

import "errors"

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

//------------------------
// Edit-buffer error codes
//------------------------

// ErrorReadonlyBuffer says buffer cannot be changed.
var ErrorReadonlyBuffer = errors.New("editbuffer.ronly")

// ErrorOldestChange says there is not more change to undo.
var ErrorOldestChange = errors.New("editbuffer.oldestChange")

// ErrorLatestChange says there is no more change to redo.
var ErrorLatestChange = errors.New("editbuffer.latestChange")

//-----
// Mode
//-----

const (
	ModeNormal byte = iota
	ModeInsert byte = iota
	ModeEx     byte = iota
	ModeVisual byte = iota
)

const Newline = `\n`
