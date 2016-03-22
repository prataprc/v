package v

type Plane struct {
	z        int
	this     *Box
	children []*Box
}

func newplane(this *Box, z int) *Plane {
	plane := &Plane{this: this, z: z}
	plane.children = make([]*Box, 0)
	return plane
}

func (plane *Plane) addbox(box *Box) {
}
