package v

import "testing"
import "fmt"

var _ = fmt.Sprintf("dummy")

func TestLayout(t *testing.T) {
	params := makeparams()
	box := NewBox("root", nil, params)
	box.Setroot(80, 40)

	box.AddBox("box1", params)
	box.AddBox("box2", params)
	box.Align()

	//box.Dump("")
}

func TestLayout(t *testing.T) {
	params := makeparams()
	box := NewBox("root", nil, params)
	box.Setroot(80, 40)

	box.AddBox("box1", params)
	box.AddBox("box2", params)
	box.Align()

	box.Dump("")
}

func makeparams() map[string]interface{} {
	return map[string]interface{}{
		"z":       0,
		"width":   0,
		"height":  -1,
		"float":   "right",
		"margin":  "1,1,1,1",
		"padding": "1,1,1,1",
	}
}
