// +build ignore

package main

import "github.com/nsf/termbox-go"

const chars = "nnnnnnnnnbbbbbbbbbuuuuuuuuuBBBBBBBBB"

func next_char(current int) int {
	current++
	if current >= len(chars) {
		return 0
	}
	return current
}

func print_combinations_table(sx, sy int, attrs []termbox.Attribute) {
	var bg termbox.Attribute
	current_char := 0
	y := sy

	all_attrs := []termbox.Attribute{
		0,
		termbox.AttrBold,
		termbox.AttrUnderline,
		termbox.AttrBold | termbox.AttrUnderline,
	}

	draw_line := func() {
		x := sx
		for _, a := range all_attrs {
			for c := termbox.ColorDefault; c <= termbox.ColorWhite; c++ {
				fg := a | c
				termbox.SetCell(x, y, rune(chars[current_char]), fg, bg)
				current_char = next_char(current_char)
				x++
			}
		}
	}

	for _, a := range attrs {
		for c := termbox.ColorDefault; c <= termbox.ColorWhite; c++ {
			bg = a | c
			draw_line()
			y++
		}
	}
}

func print_wide(x, y int, s string) {
	red := false
	for _, r := range s {
		c := termbox.ColorDefault
		if red {
			c = termbox.ColorRed
		}
		termbox.SetCell(x, y, r, termbox.ColorDefault, c)
		x += 2
		red = !red
	}
}

const hello_world = "こんにちは世界"

func draw_title() {
	title := "COLOR CODES"
}

func draw_all() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	switch output_mode {
	case termbox.OutputNormal:
		print_combinations_table(1, 1, []termbox.Attribute{
			0,
			termbox.AttrBold,
		})
		print_combinations_table(2+len(chars), 1, []termbox.Attribute{
			termbox.AttrReverse,
		})
		print_wide(2+len(chars), 11, hello_world)

	case termbox.OutputGrayscale:
		for y := 0; y < 24; y++ {
			for x := 0; x < 24; x++ {
				termbox.SetCell(x, y, 'n',
					termbox.Attribute(x+1),
					termbox.Attribute(y+1))
				termbox.SetCell(x+25, y, 'b',
					termbox.Attribute(x+1)|termbox.AttrBold,
					termbox.Attribute(24-y))
				termbox.SetCell(x+50, y, 'u',
					termbox.Attribute(x+1)|termbox.AttrUnderline,
					termbox.Attribute(y+1))
			}
			termbox.SetCell(76, y, 'd',
				termbox.Attribute(y+1),
				termbox.ColorDefault)
			termbox.SetCell(77, y, 'd',
				termbox.ColorDefault,
				termbox.Attribute(24-y))
		}

	case termbox.Output216:
		for r := 0; r < 6; r++ {
			for g := 0; g < 6; g++ {
				for b := 0; b < 6; b++ {
					y := r
					x := g + 6*b
					c1 := termbox.Attribute(1 + r*36 + g*6 + b)
					bg := termbox.Attribute(1 + g*36 + b*6 + r)
					c2 := termbox.Attribute(1 + b*36 + r*6 + g)
					bc1 := c1 | termbox.AttrBold
					uc1 := c1 | termbox.AttrUnderline
					bc2 := c2 | termbox.AttrBold
					uc2 := c2 | termbox.AttrUnderline
					termbox.SetCell(x, y, 'n', c1, bg)
					termbox.SetCell(x, y+6, 'b', bc1, bg)
					termbox.SetCell(x, y+12, 'u', uc1, bg)
					termbox.SetCell(x, y+18, 'B', bc1|uc1, bg)
					termbox.SetCell(x+37, y, 'n', c2, bg)
					termbox.SetCell(x+37, y+6, 'b', bc2, bg)
					termbox.SetCell(x+37, y+12, 'u', uc2, bg)
					termbox.SetCell(x+37, y+18, 'B', bc2|uc2, bg)
				}
				c1 := termbox.Attribute(1 + g*6 + r*36)
				c2 := termbox.Attribute(6 + g*6 + r*36)
				termbox.SetCell(74+g, r, 'd', c1, termbox.ColorDefault)
				termbox.SetCell(74+g, r+6, 'd', c2, termbox.ColorDefault)
				termbox.SetCell(74+g, r+12, 'd', termbox.ColorDefault, c1)
				termbox.SetCell(74+g, r+18, 'd', termbox.ColorDefault, c2)
			}
		}

	case termbox.Output256:
		for y := 0; y < 4; y++ {
			for x := 0; x < 8; x++ {
				for z := 0; z < 8; z++ {
					bg := termbox.Attribute(1 + y*64 + x*8 + z)
					c1 := termbox.Attribute(256 - y*64 - x*8 - z)
					c2 := termbox.Attribute(1 + y*64 + z*8 + x)
					c3 := termbox.Attribute(256 - y*64 - z*8 - x)
					c4 := termbox.Attribute(1 + y*64 + x*4 + z*4)
					bold := c2 | termbox.AttrBold
					under := c3 | termbox.AttrUnderline
					both := c1 | termbox.AttrBold | termbox.AttrUnderline
					termbox.SetCell(z+8*x, y, ' ', 0, bg)
					termbox.SetCell(z+8*x, y+5, 'n', c4, bg)
					termbox.SetCell(z+8*x, y+10, 'b', bold, bg)
					termbox.SetCell(z+8*x, y+15, 'u', under, bg)
					termbox.SetCell(z+8*x, y+20, 'B', both, bg)
				}
			}
		}
		for x := 0; x < 12; x++ {
			for y := 0; y < 2; y++ {
				c1 := termbox.Attribute(233 + y*12 + x)
				termbox.SetCell(66+x, y, 'd', c1, termbox.ColorDefault)
				termbox.SetCell(66+x, 2+y, 'd', termbox.ColorDefault, c1)
			}
		}
		for x := 0; x < 6; x++ {
			for y := 0; y < 6; y++ {
				c1 := termbox.Attribute(17 + x*6 + y*36)
				c2 := termbox.Attribute(17 + 5 + x*6 + y*36)
				termbox.SetCell(66+x, 6+y, 'd', c1, termbox.ColorDefault)
				termbox.SetCell(66+x, 12+y, 'd', c2, termbox.ColorDefault)
				termbox.SetCell(72+x, 6+y, 'd', termbox.ColorDefault, c1)
				termbox.SetCell(72+x, 12+y, 'd', termbox.ColorDefault, c2)
			}
		}

	}

	termbox.Flush()
}

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	termbox.SetOutputMode(termbox.Output256)

	draw_all()
loop:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc:
				break loop
			case termbox.KeyArrowUp, termbox.KeyArrowRight:
				switch_output_mode(1)
				draw_all()
			case termbox.KeyArrowDown, termbox.KeyArrowLeft:
				switch_output_mode(-1)
				draw_all()
			}
		case termbox.EventResize:
			draw_all()
		}
	}
}
