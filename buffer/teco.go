// +build ignore

package buffer

type Command struct {
	name string
	ebuf *EditBuffer
	args []interface{}
}

var TecoNormals = map[string]func(Command){
//"h"    : h,
//"<-"   : h,
//"<bs>" : bs,
}

// horizontal movement, returns new cursor position.
func h(iter LineIterator, dot, distance int64) int64 {
	return 0
}
