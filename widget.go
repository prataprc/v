package v

import term "github.com/prataprc/v/term"

type Widget interface {
	Render()
}

type Lineiterator interface {
	Next() []term.Cell
}
