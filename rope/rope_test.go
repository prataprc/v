package rope

import "testing"
import "io/ioutil"
import "fmt"
import "math/rand"

var _ = fmt.Sprintf("dummy")

var testRopeBufferCapacity = int64(10 * 1024)

func TestRopeSample256(t *testing.T) {
	data, err := ioutil.ReadFile("../testdata/sample.txt")
	if err != nil {
		t.Fatal(err)
	}
	rb := NewRopebuffer([]rune(string(data)), 256)
	stats := rb.Stats()
	length := validateRopeBuild(t, stats)
	if l := int64(len(data)); length != l {
		t.Fatalf("mismatch in length %v, got %v", length, l)
	}
}

func TestRopeSample1MB(t *testing.T) {
	data, err := ioutil.ReadFile("../testdata/sample.txt")
	if err != nil {
		t.Fatal(err)
	}
	rb := NewRopebuffer([]rune(string(data)), 1024*1024)
	stats := rb.Stats()
	length := validateRopeBuild(t, stats)
	if l := int64(len(data)); length != l {
		t.Fatalf("mismatch in length %v, got %v", length, l)
	}
}

func TestRopeIndex(t *testing.T) {
	rb := NewRopebuffer([]rune("hello world"), 2)
	if err := validateRead(rb); err != nil {
		t.Fatal(err)
	}
}

func TestRopeSubstr(t *testing.T) {
	rb := NewRopebuffer([]rune("hello world"), 2)
	if err := validateSubstr(rb); err != nil {
		t.Fatal(err)
	}
}

func TestRopeDicing(t *testing.T) {
	rb := NewRopebuffer([]rune("hello world"), 2)
	if err := validateDicing(rb); err != nil {
		t.Fatal(err)
	}
}

func TestRopeInsert(t *testing.T) {
	var err error

	rb := NewRopebuffer([]rune("hello world"), 2)
	lb := NewLinearBuffer([]rune("hello world"))
	lenrb := len(rb.Value())
	insVals := [][]rune{[]rune(""), []rune("a"), []rune("alpha")}
	for i := 0; i < 100; i++ {
		dot := int64(rand.Intn(lenrb))
		for _, insVal := range insVals {
			rb, err = rb.Insert(dot, insVal, true)
			if err != nil {
				t.Fatal(err)
			}
			lb, err = lb.Insert(dot, insVal, true)
			if err != nil {
				t.Fatal(err)
			}
			if err = validateRead(rb); err != nil {
				t.Fatal(err)
			}
			if err = validateDicing(rb); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func BenchmarkRopeSample8(b *testing.B) {
	data, err := ioutil.ReadFile("../testdata/sample.txt")
	if err != nil {
		b.Fatal(err)
	}
	runes := []rune(string(data))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewRopebuffer(runes, 8)
	}
	b.SetBytes(int64(len(data)))
}

func BenchmarkRopeSample256(b *testing.B) {
	data, err := ioutil.ReadFile("../testdata/sample.txt")
	if err != nil {
		b.Fatal(err)
	}
	runes := []rune(string(data))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewRopebuffer(runes, 256)
	}
	b.SetBytes(int64(len(data)))
}

func BenchmarkRopeLength(b *testing.B) {
	data, err := ioutil.ReadFile("../testdata/sample.txt")
	if err != nil {
		b.Fatal(err)
	}
	runes := []rune(string(data))
	rb := NewRopebuffer(runes, testRopeBufferCapacity)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb.Length()
	}
}

func BenchmarkRopeValue(b *testing.B) {
	data, err := ioutil.ReadFile("../testdata/sample.txt")
	if err != nil {
		b.Fatal(err)
	}
	runes := []rune(string(data))
	rb := NewRopebuffer(runes, testRopeBufferCapacity)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb.Value()
	}
}

func BenchmarkRopeIndex(b *testing.B) {
	data, err := ioutil.ReadFile("../testdata/sample.txt")
	if err != nil {
		b.Fatal(err)
	}
	runes := []rune(string(data))
	rb := NewRopebuffer(runes, testRopeBufferCapacity)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb.Index(1024 * 512)
	}
}

func BenchmarkRopeSubstr(b *testing.B) {
	data, err := ioutil.ReadFile("../testdata/sample.txt")
	if err != nil {
		b.Fatal(err)
	}
	runes := []rune(string(data))
	rb := NewRopebuffer(runes, testRopeBufferCapacity)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb.Substr(1024*512, 10000)
	}
}

func BenchmarkRopeInsert(b *testing.B) {
	data, err := ioutil.ReadFile("../testdata/sample.txt")
	if err != nil {
		b.Fatal(err)
	}
	runes := []rune(string(data))
	rb := NewRopebuffer(runes, testRopeBufferCapacity)
	s, err := rb.Substr(0, 10000)
	if err != nil {
		b.Fatal(err)
	}
	runes = []rune(s)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb.Insert(1024*512, runes, true)
	}
}

func BenchmarkBytes2Str(b *testing.B) {
	var text string
	data, err := ioutil.ReadFile("./rope_test.go")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		text = string(data)
	}
	b.SetBytes(int64(len(text)))
}

func BenchmarkStr2Runes(b *testing.B) {
	var text []rune
	data, err := ioutil.ReadFile("./rope_test.go")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		text = []rune(string(data))
	}
	b.SetBytes(int64(len(text)))
}

func validateRopeBuild(t *testing.T, stats map[string]interface{}) int64 {
	if deviant := stats["deviantLevel"].(float64); deviant != 0.0 {
		t.Fatalf("expected deviant level 0, got %v", deviant)
	}
	if minl := stats["minLevel"].(int64); minl != 1 {
		t.Fatalf("expected minLevel as 1, got %v", minl)
	}
	maxLevel := stats["maxLevel"].(int64)
	if avg := stats["meanLevel"].(float64); float64(maxLevel) != avg {
		t.Fatalf("unexpected meanLevel %v, got %v", maxLevel, avg)
	}
	length, capacity := stats["capacity"].(int64), stats["length"].(int64)
	if length != capacity {
		t.Fatalf("unexpected length/capacity: %v/%v", length, capacity)
	}
	leafs := stats["leafs"].(int64)
	if ref := 2 << (uint64(maxLevel) - 2); int64(ref) != leafs {
		t.Fatalf("unexpected leafs %v, got %v", ref, leafs)
	}
	nodes := stats["nodes"].(int64)
	if ref := (2 << (uint64(maxLevel) - 2)) - 1; int64(ref) != nodes {
		t.Fatalf("unexpected leafs %v, got %v", ref, nodes)
	}
	return length
}

func validateRead(rb *RopeBuffer) error {
	lb := NewLinearBuffer(rb.Value())
	ref := lb.Value()
	if x, y := string(rb.Value()), string(ref); x != y {
		return fmt.Errorf("mismatch in value %q, got %q", x, y)
	}

	length, err := rb.Length()
	if err != nil {
		return err
	}
	if length == 0 {
		if _, ok, err := rb.Index(0); err == nil || ok == true {
			return fmt.Errorf("expecting error")
		}
		if _, ok, err := rb.Index(1); err == nil || ok == true {
			return fmt.Errorf("expecting error")
		}
	} else {
		for _, dot := range []int64{0, length / 2, length - 1} {
			if ch, ok, err := rb.Index(dot); err != nil {
				return err
			} else if ok == false {
				return fmt.Errorf("ok is false for dot %v ch %v", dot, ref[dot])
			} else if ch != ref[dot] {
				return fmt.Errorf("mismatch for ch %v, got %v", ref[dot], ch)
			}
		}
		if _, ok, err := rb.Index(length); err == nil || ok == true {
			return fmt.Errorf("expecting error for %v", length)
		}
	}

	return validateSubstr(rb)
}

func validateDicing(rbRef *RopeBuffer) error {
	x := rbRef.Value()
	lenx := int64(len(x))
	for _, dot := range []int64{0, lenx/2 - 1, lenx/2 + 2, lenx / 2, lenx - 1} {
		rbLeft, rbRight, err := rbRef.Split(dot)
		if err != nil {
			fmt.Errorf("failed splitting %q at %d: %v", string(x), dot, err)
		}
		rb, err := rbLeft.Concat(rbRight)
		if err != nil {
			x, y := string(rbLeft.Value()), string(rbRight.Value())
			fmt.Errorf("failed concating %q and %q: %v", x, y, err)
		}
		if err := validateRead(rb); err != nil {
			fmt.Errorf("validateRead() %q at %d: %v", string(x), dot, err)
		}
	}
	_, _, err := rbRef.Split(lenx)
	if err == nil {
		fmt.Errorf("expecting error splitting %q at %d", string(x), lenx)
	}
	return nil
}

func validateSubstr(rbRef *RopeBuffer) error {
	x := rbRef.Value()
	lenx := len(x)

	lb := NewLinearBuffer(x)
	argList := make([][2]int64, 0)
	for dot := 0; dot < lenx; dot++ {
		for n := 0; n < (dot - lenx); n++ {
			argList = append(argList, [2]int64{int64(dot), int64(n)})
		}
	}
	for _, arg := range argList {
		dot, n := arg[0], arg[1]
		a, err := rbRef.Substr(dot, n)
		if err != nil {
			return fmt.Errorf("rb at %d size %d for %q", dot, n, string(x))
		}
		b, err := lb.Substr(dot, n)
		if err != nil {
			return fmt.Errorf("lb at %d size %d for %q", dot, n, string(x))
		}
		if a != b {
			return fmt.Errorf("at %d size %d for %q", dot, n, string(x))
		}
	}
	return nil
}
