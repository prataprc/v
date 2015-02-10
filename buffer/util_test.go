package buffer

import "testing"
import "fmt"
import "unicode/utf8"

var _ = fmt.Sprintf("dummy")

var testChinese = `左司馬販（《春秋左傳·哀公四年》 #2.2）
當為左司馬「眅」，音「攀」上聲。 
並非「販賣」之「販」`

func TestGetRuneStart(t *testing.T) {
	var err error

	offs := make([]int64, 0)
	for i := int64(0); i < int64(len(testChinese)); i++ {
		n, err := getRuneStart([]byte(testChinese[i:]), false)
		if err != nil {
			break
		}
		i += n
		offs = append(offs, i)
	}
	runes := make([]rune, len(offs))
	for i, off := range offs {
		runes[i], _ = utf8.DecodeRune([]byte(testChinese[off:]))
		if err != nil {
			t.Fatal(err)
		}
	}
	if testChinese != string(runes) {
		t.Fatalf("mismatch %v, got %v\n", testChinese, string(runes))
	}
}

func TestGetRuneStartReverse(t *testing.T) {
	var err error

	offs := make([]int64, 0)
	for n := int64(len(testChinese)); n > 0; {
		n, err = getRuneStart([]byte(testChinese[:n]), true)
		if err != nil {
			break
		}
		offs = append(offs, n)
	}

	runes := make([]rune, len(offs))
	for i, off := range offs {
		runes[i], _ = utf8.DecodeRune([]byte(testChinese[off:]))
		if err != nil {
			t.Fatal(err)
		}
	}
	l := len(runes) - 1
	for i := 0; i <= l/2; i++ {
		runes[i], runes[l-i] = runes[l-i], runes[i]
	}
	if testChinese != string(runes) {
		t.Fatalf("mismatch %v, got %v\n", testChinese, string(runes))
	}
}

func TestBytes2RunesN(t *testing.T) {
	refrn := int64(len([]rune(testChinese)))
	acc := make([]rune, refrn)
	rn, size, err := bytes2RunesN([]byte(testChinese), refrn, acc)
	if err != nil {
		t.Fatal(err)
	} else if rn != refrn {
		t.Fatalf("expected %v runes, got %v\n", refrn, rn)
	} else if size != int64(len(testChinese)) {
		t.Fatalf("expected %v bytes, got %v\n", len(testChinese), size)
	} else if string(acc) != testChinese {
		t.Fatalf("expected %v\ngot %v\n", testChinese, string(acc))
	}
}

func TestBytes2Runes(t *testing.T) {
	runes, err := bytes2Runes([]byte(testChinese))
	if err != nil {
		t.Fatal(err)
	} else if string(runes) != testChinese {
		t.Fatalf("expected %v\ngot %v\n", testChinese, string(runes))
	}
}

func BenchmarkGetRuneStart(b *testing.B) {
	bytes := []byte(testChinese)
	for j := 0; j < b.N; j++ {
		for i := int64(0); i < int64(len(testChinese)); i++ {
			n, _ := getRuneStart(bytes[i:], false)
			i += n
		}
	}
	b.SetBytes(int64(len(testChinese)))
}

func BenchmarkGetRuneStartRev(b *testing.B) {
	bytes := []byte(testChinese)
	for j := 0; j < b.N; j++ {
		for n := int64(len(bytes)); n > 0; {
			n, _ = getRuneStart([]byte(bytes[:n]), true)
		}
	}
	b.SetBytes(int64(len(testChinese)))
}

func BenchmarkBytes2RunesN(b *testing.B) {
	bytes := []byte(testChinese)
	refrn := int64(len([]rune(testChinese)))
	acc := make([]rune, refrn)
	for i := 0; i < b.N; i++ {
		bytes2RunesN(bytes, refrn, acc)
	}
	b.SetBytes(int64(len(testChinese)))
}

func BenchmarkBytes2Runes(b *testing.B) {
	bytes := []byte(testChinese)
	for i := 0; i < b.N; i++ {
		bytes2Runes(bytes)
	}
	b.SetBytes(int64(len(testChinese)))
}
