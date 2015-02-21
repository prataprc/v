package buffer

import "testing"
import "fmt"

var _ = fmt.Sprintf("dummy")

func TestEditBufferLines(t *testing.T) {
	//s := `\n\n左司\n馬販（《\n春秋左\n\n傳·哀\n公四年\n》 當為左\n\n\n司\n\n`
	//makeEbuf := func(s string) *EditBuffer {
	//    rb, err := NewRopebuffer([]byte(s), 4)
	//    if err != nil {
	//        t.Fatal(err)
	//    }
	//    return NewEditBuffer(0, rb, nil)
	//}
	//ebuf := makeEbuf(``)
	//fmt.Println(ebuf.Lines(-1, 1))
}
