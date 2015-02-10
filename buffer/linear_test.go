package buffer

import "testing"
import "io"
import "fmt"

var _ = fmt.Sprintf("dummy")

func TestLinearStreamFrom(t *testing.T) {
	var err error
	var r rune
	var size int
	lb := NewLinearBuffer([]byte(testChinese))
	reader := lb.StreamFrom(0)
	totalrn, totalsize, runes := 0, 0, make([]rune, len([]rune(testChinese)))
	for err != io.EOF {
		r, size, err = reader.ReadRune()
		if err != io.EOF {
			runes[totalrn] = r
			totalrn++
			totalsize += size
		}
	}
	if l := len([]rune(testChinese)); totalrn != l {
		t.Fatalf("rune len expected %v, got %v\n", l, totalrn)
	} else if l := len(testChinese); totalsize != l {
		t.Fatalf("size expected %v, got %v\n", totalsize, l)
	} else if s := string(runes); s != testChinese {
		t.Fatalf("expected %v\ngot %v\n", testChinese, s)
	}
}

func BenchmarkLinearStreamFrom(b *testing.B) {
	var err error
	lb := NewLinearBuffer([]byte(testChinese))
	reader := lb.StreamFrom(0)
	for i := 0; i < b.N; i++ {
		for err != io.EOF {
			_, _, err = reader.ReadRune()
		}
	}
	b.SetBytes(int64(len(testChinese)))
}
