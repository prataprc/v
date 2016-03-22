package v

import "strings"
import "strconv"
import "fmt"

import term "github.com/prataprc/v/term"

var Maxplanes = 10

// Box is where the buffer is rendered.
type Box struct {
	name      string
	container *Box
	planes    []*Plane
	buffer    Buffer
	display   term.Cell

	// properties
	x, y              int          // absolute reference from (1,1)
	width, height     int          // border to border, includes padding
	margins, paddings []int        // top, right, down, left
	borders           []*term.Cell // top, right, down, left
}

// params
// `width` - width of the box
// `height` - height of the box
// `border` - border specification for all sides
// `padding` - padding specification for all sides
func NewBox(name string, container *Box, params map[string]interface{}) *Box {
	box := &Box{name: name, container: container}

	// width and height
	contwidth, contheight := container.planesize()
	box.width = box.getparam(params, "width", contwidth).(int)
	box.height = box.getparam(params, "height", contheight).(int)

	// outline
	box.margins = box.parammargins(params)
	box.borders = box.paramborders(params)
	box.paddings = box.parampaddings(params)

	// planes and display buffer
	planes := make([]*Plane, 0, Maxplanes)
	for z := 0; z < Maxplanes; z++ {
		planes = append(planes, newplane(box, z))
	}
	box.planes = planes
	return box
}

func (box *Box) AddBox(name string, params map[string]interface{}) {
	cbox := NewBox(name, box, params)
	z := box.getparam(params, "z", 0).(int)
	box.planes[z].addbox(cbox)
}

func (box *Box) planesize() (width, height int) {
	width = box.width - box.paddings[1] - box.paddings[3]
	height = box.width - box.paddings[0] - box.paddings[2]
	return width, height
}

func (box *Box) planexy() (x, y int) {
	return box.paddings[0], box.paddings[3]
}

func (box *Box) bordercells() {
}

func (box *Box) paddingcells() {
}

// parse parameters

func (box *Box) parammargins(params map[string]interface{}) []int {
	arg, _ := params["margin"]
	rv := make([]int, 0)
	switch margin := arg.(type) {
	case string:
		margin = strings.Trim(margin, " \t\r\n")
		for _, item := range strings.Split(margin, ",") {
			n, err := strconv.Atoi(item)
			if err != nil {
				panic(fmt.Errorf("box %q invalid padding: %v", box.name, err))
			}
			rv = append(rv, n)
		}
	case []int:
		rv = append(rv, margin...)
	default:
		rv = append(rv, 0, 0, 0, 0)
	}
	switch len(rv) {
	case 1:
		return append(rv, rv[0], rv[0], rv[0])
	case 2:
		return append(rv, rv[0], rv[1])
	case 4:
		return rv
	}
	panic(fmt.Errorf("box %v, invalid number of margins: %v", box.name, rv))
}

func (box *Box) parampaddings(params map[string]interface{}) []int {
	arg, _ := params["padding"]
	rv := make([]int, 0)
	switch padd := arg.(type) {
	case string:
		padd = strings.Trim(padd, " \t\r\n")
		for _, item := range strings.Split(padd, ",") {
			n, err := strconv.Atoi(item)
			if err != nil {
				panic(fmt.Errorf("box %q invalid padding: %v", box.name, err))
			}
			rv = append(rv, n)
		}
	case []int:
		rv = append(rv, padd...)
	default:
		rv = append(rv, 0, 0, 0, 0)
	}
	switch len(rv) {
	case 1:
		return append(rv, rv[0], rv[0], rv[0])
	case 2:
		return append(rv, rv[0], rv[1])
	case 4:
		return rv
	}
	panic(fmt.Errorf("box %v, invalid number of margins: %v", box.name, rv))
}

var BorderLine = [4]rune{'─', '│', '─', '│'}

//	'∗'

func (box *Box) paramborders(p map[string]interface{}) []*term.Cell {
	item, ok := p["border"]
	if !ok {
		return nil
	}
	borders := strings.Trim(item.(string), " \t\r\n")

	cells := make([]*term.Cell, 0)
	for i, arg := range strings.Split(borders, ";") {
		cells = append(cells, box.paramborder(i, arg))
	}

	if len(cells) == 4 {
		return cells
	}
	panic(fmt.Errorf("box %v, specify all borders"))
}

func (box *Box) paramborder(side int, border string) (c *term.Cell) {
	var fgok bool
	for _, arg := range strings.Split(strings.Trim(border, " \t\r\n"), ",") {
		switch arg {
		case "line":
			c = new(term.Cell)
			c.Ch = BorderLine[side]

		case "none":
			return nil

		default:
			if c == nil {
				panic(fmt.Errorf("box %q start with border type", box.name))
			} else if fgok == false {
				c.Fg = box.paramattr(arg)
				fgok = true
			}
			c.Bg = box.paramattr(arg)
		}
	}
	return
}

func (box *Box) paramattr(attr string) (a term.Attribute) {
	for _, arg := range strings.Split(attr, ":") {
		switch arg {
		case "black":
			a |= term.ColorBlack
		case "red":
			a |= term.ColorRed
		case "green":
			a |= term.ColorGreen
		case "yellow":
			a |= term.ColorYellow
		case "blue":
			a |= term.ColorBlue
		case "magenta":
			a |= term.ColorMagenta
		case "cyan":
			a |= term.ColorCyan
		case "white":
			a |= term.ColorWhite
		case "bold":
			a |= term.AttrBold
		case "underline":
			a |= term.AttrUnderline
		case "reverse":
			a |= term.AttrReverse
		default:
			val, err := strconv.Atoi(arg)
			if err != nil {
				panic(fmt.Errorf("box %q, attribute error: %v", box.name, err))
			}
			a |= term.Attribute(val)
		}
	}
	return a
}

func (box *Box) getparam(
	params map[string]interface{}, key string, def interface{}) interface{} {

	if val, ok := params[key]; ok {
		return val
	}
	return def
}
