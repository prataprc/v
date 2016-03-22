package v

import term "github.com/prataprc/v/term"

type Buffer interface {
	Rangelines(from, to int) Lineiterator
}

type Lineiterator interface {
	Next() []term.Cell
}
