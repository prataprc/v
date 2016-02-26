package v

import "fmt"

type Dimension interface {
	Size() (width, height int)
	Margin() (top, right, bottom, left int)
	Border() (top, right, bottom, left int)
	Padding() (top, right, bottom, left int)
	Float() (side string)
	Setcoordinate(x, y int)
}

type packbox struct {
	box                 *Box
	x, y, width, height int
}

func newpackbox(x, y, width, height int) *packbox {
	return &packbox{x: x, y: y, width: width, height: height}
}

func (pb *packbox) fit(boxes []*Box) []*Box {
	var err error

	if pb == nil {
		return boxes
	} else if len(boxes) == 0 {
		return nil
	}

	box := boxes[0]
	margins, err := box.fixmargins(pb.width)
	if err != nil {
		panic(fmt.Errorf("error fixing margins: %v", err))
	}
	paddings, err := box.fixpaddings(pb.width)
	if err != nil {
		panic(fmt.Errorf("error fixing margins: %v", err))
	}
	mt, mr, mb, ml := margins[0], margins[1], margins[2], margins[3]

	width, height := box.Size()
	side := box.Float()
	if width < 0 { // minimum width
		if boxwidth := width - mr - ml; pb.width < (-boxwidth) {
			return boxes
		}
		width = pb.width - mr - ml
	} else if width == 0 {
		width = pb.width - mr - ml
	}
	fullw := width + mr + ml

	if width > 0 && pb.width >= fullw {
		box.margins, box.paddings = margins, paddings

		var left, bottom, right *packbox
		options := make([]*packbox, 0)
		bottom = newpackbox(pb.x, pb.y+height+mt+mb, pb.width, -1)
		switch side {
		case "left":
			right = newpackbox(pb.x+fullw, pb.y, pb.width-fullw, -1)
			box.Setcoordinate(pb.x+ml, pb.y+mt)
			box.Setsize(width, height)
			options = append(options, right, bottom)
		case "right":
			left = newpackbox(pb.x, pb.y, pb.width-fullw, -1)
			box.Setcoordinate(pb.x+(pb.width-fullw)+ml, pb.y+mt)
			box.Setsize(width, height)
			options = append(options, left, bottom)
		}
		boxes = boxes[1:]
		for _, option := range options {
			boxes = option.fit(boxes)
		}
		return boxes
	}
	return boxes
}
