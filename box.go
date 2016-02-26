package v

import "strings"
import "strconv"
import "fmt"

import term "github.com/prataprc/v/term"

var Maxplanes = 10

// Box is where the buffer is rendered.
type Box struct {
	name       string
	container  *Box
	containerz int
	planes     []*Plane
	widget     Widget
	float      string

	// properties
	x, y                int          // relative to plane, excludes padding
	width, height       int          // excludes border, includes padding
	tmargins, tpaddings []string     // top, right, down, left
	margins, paddings   []int        // top, right, down, left
	bordercells         []*term.Cell // top, right, down, left
	borders             []int        // top, right, down, left
}

type Plane struct {
	z        int
	this     *Box
	children []*Box
}

// params
// `z` - stack level in zaxis
// `width` - width of the box
// `height` - height of the box
// `float` - affinity to float, "left", "right"
// `display` - whether to display the box or not
// `margin` - margin specification for all sides
// `border` - border specification for all sides
// `padding` - padding specification for all sides
func NewBox(name string, container *Box, params map[string]interface{}) *Box {
	z := boxgetparam(params, "z", 0).(int)
	float := boxgetparam(params, "float", "left").(string)
	box := &Box{
		name: name, container: container, containerz: z, float: float,
	}
	// planes and display buffer
	planes := make([]*Plane, 0, Maxplanes)
	for z := 0; z < Maxplanes; z++ {
		plane := &Plane{this: box, z: z, children: make([]*Box, 0)}
		planes = append(planes, plane)
	}
	box.planes = planes

	height := box.Root().height

	// outline
	box.tmargins = box.parsemargins(params)
	box.tpaddings = box.parsepaddings(params)
	box.bordercells, box.borders = box.parseborders(params)
	box.width = boxgetparam(params, "width", 0).(int)
	box.height = boxgetparam(params, "height", height).(int)
	if box.height < 0 {
		box.height = height
	}

	return box
}

func (box *Box) Root() *Box {
	if box.container == nil {
		return box
	}
	return box.container.Root()
}

func (box *Box) String() string {
	fmsg := "box#%v{(%v,%v) -%v- |%v| m:%v, b:%v, p:%v}"
	return fmt.Sprintf(
		fmsg,
		box.name, box.x, box.y, box.width, box.height, box.margins,
		box.borders, box.paddings)
}

func (box *Box) Setroot(contw, conth int) {
	var err error
	if box.margins, err = box.fixmargins(contw); err != nil {
		panic(fmt.Errorf("error fixing margins: %v", err))
	}
	if box.paddings, err = box.fixpaddings(contw); err != nil {
		panic(fmt.Errorf("error fixing padding: %v", err))
	}
	box.x, box.y = box.margins[3], box.margins[0]
	box.width = contw - box.margins[1] - box.margins[3]
	box.height = conth - box.margins[0] - box.margins[2]
}

func (box *Box) AddBox(name string, params map[string]interface{}) {
	child := NewBox(name, box, params)
	z := boxgetparam(params, "z", 0).(int)
	plane := box.planes[z]
	plane.children = append(plane.children, child)
}

func (box *Box) Align() {
	x := box.x + box.borders[3] + box.paddings[3]
	y := box.y + box.borders[0] + box.paddings[0]
	width := box.width - box.borders[1] - box.borders[3]
	width -= (box.paddings[1] + box.paddings[3])
	for _, plane := range box.planes {
		root := newpackbox(x, y, width, -1)
		if len(root.fit(plane.children)) != 0 {
			panic("unable to fix")
		}
	}
}

func (box *Box) Setsize(width, height int) {
	box.width, box.height = width, height
}

func (box *Box) Dump(prefix string) {
	fmt.Printf("%v%v\n", prefix, box)
	for i, plane := range box.planes {
		if len(plane.children) > 0 {
			fmt.Printf("%vPlane: %v\n", prefix, i)
			for _, box := range plane.children {
				box.Dump(prefix + "  ")
			}
		}
	}
}

//---- Dimension{} interface

func (box *Box) Size() (width, height int) {
	return box.width, box.height
}

func (box *Box) Margin() (top, right, bottom, left int) {
	margins := box.margins
	return margins[0], margins[1], margins[2], margins[3]
}

func (box *Box) Border() (top, right, bottom, left int) {
	borders := box.borders
	return borders[0], borders[1], borders[2], borders[3]
}

func (box *Box) Padding() (top, right, bottom, left int) {
	paddings := box.paddings
	return paddings[0], paddings[1], paddings[2], paddings[3]
}

func (box *Box) Float() (side string) {
	return box.float
}

func (box *Box) Setcoordinate(x, y int) {
	box.x, box.y = x, y
}

//---- local functions

func (box *Box) parsemargins(params map[string]interface{}) []string {
	value, ok := params["margin"]
	if !ok {
		return []string{"0", "0", "0", "0"}
	}

	margin := strings.Trim(value.(string), " \t\r\n")
	rv := make([]string, 0)
	for _, item := range strings.Split(margin, ",") {
		rv = append(rv, item)
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

func (box *Box) parsepaddings(params map[string]interface{}) []string {
	value, ok := params["padding"]
	if !ok {
		return []string{"0", "0", "0", "0"}
	}

	padding := strings.Trim(value.(string), " \t\r\n")
	rv := make([]string, 0)
	for _, item := range strings.Split(padding, ",") {
		rv = append(rv, item)
	}
	switch len(rv) {
	case 1:
		return append(rv, rv[0], rv[0], rv[0])
	case 2:
		return append(rv, rv[0], rv[1])
	case 4:
		return rv
	}
	panic(fmt.Errorf("box %v, invalid number of paddings: %v", box.name, rv))
}

func (box *Box) fixmargins(contw int) ([]int, error) {
	margins := make([]int, 0)
	for _, item := range box.tmargins {
		var n int
		if strings.HasSuffix(item, "%") {
			f, err := strconv.ParseFloat(item[:len(item)-1], 64)
			if err != nil {
				return nil, err
			} else {
				n = int(float64(contw) * (f / 100))
			}
		}
		f, err := strconv.ParseFloat(item, 64)
		if err != nil {
			return nil, err
		}
		n = int(f)
		margins = append(margins, n)
	}
	return margins, nil
}

func (box *Box) fixpaddings(contw int) ([]int, error) {
	paddings := make([]int, 0)
	for _, item := range box.tpaddings {
		var n int
		if strings.HasSuffix(item, "%") {
			f, err := strconv.ParseFloat(item[:len(item)-1], 64)
			if err != nil {
				return nil, err
			} else {
				n = int(float64(contw) * (f / 100))
			}
		}
		f, err := strconv.ParseFloat(item, 64)
		if err != nil {
			return nil, err
		}
		n = int(f)
		paddings = append(paddings, n)
	}
	return paddings, nil
}

var BorderLine = [4]rune{'─', '│', '─', '│'}

func (box *Box) parseborders(p map[string]interface{}) ([]*term.Cell, []int) {
	item, ok := p["border"]
	if !ok {
		return []*term.Cell{nil, nil, nil, nil}, []int{0, 0, 0, 0}
	}
	border := strings.Trim(item.(string), " \t\r\n")

	cells, borders := make([]*term.Cell, 0), make([]int, 4)
	for i, arg := range strings.Split(border, ";") {
		cell := box.parseborder(i, arg)
		cells = append(cells, cell)
		borders[i] = 1
		if cell == nil {
			borders[i] = 0
		}
	}
	if len(cells) == 4 {
		return cells, borders
	}
	panic(fmt.Errorf("box %v, specify all borders"))
}

// <type>[,<color:attribute>,<color:attribute>]
func (box *Box) parseborder(side int, border string) (c *term.Cell) {
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
				panic(fmt.Errorf("box %q, start with border type", box.name))
			} else if fgok == false {
				fgok = true
				c.Fg = box.parsebrdrattr(arg)
			}
			c.Bg = box.parsebrdrattr(arg)
		}
	}
	return
}

// <color:attribute>
func (box *Box) parsebrdrattr(attr string) (a term.Attribute) {
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

func boxgetparam(
	params map[string]interface{}, key string, def interface{}) interface{} {

	if val, ok := params[key]; ok {
		return val
	}
	return def
}
