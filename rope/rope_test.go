package rope

import "testing"
import "io/ioutil"
import "fmt"
import "math/rand"

import "github.com/prataprc/v"

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
	if err := validateRead(rb, nil); err != nil {
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

func TestRopeInsertBasic(t *testing.T) {
	rb := NewRopebuffer([]rune("hello world"), 2)
	// before begin
	if _, err := rb.Insert(-1, []rune("a")); err != v.ErrorIndexOutofbound {
		t.Fatalf("expecting err ErrorIndexOutofbound")
	} else if err = validateRead(rb, []rune("hello world")); err != nil {
		t.Fatal(err)
		// at begin
	} else if rb, err = rb.Insert(0, []rune("1")); err != nil {
		t.Fatal(err)
	} else if err = validateRead(rb, []rune("1hello world")); err != nil {
		t.Fatal(err)
		// before middle
	} else if rb, err = rb.Insert(5, []rune("2")); err != nil {
		t.Fatal(err)
	} else if err = validateRead(rb, []rune("1hell2o world")); err != nil {
		t.Fatal(err)
		// after middle
	} else if rb, err = rb.Insert(7, []rune("3")); err != nil {
		t.Fatal(err)
	} else if err = validateRead(rb, []rune("1hell2o3 world")); err != nil {
		t.Fatal(err)
		// at middle
	} else if rb, err = rb.Insert(8, []rune("4")); err != nil {
		t.Fatal(err)
	} else if err = validateRead(rb, []rune("1hell2o34 world")); err != nil {
		t.Fatal(err)
		// at end
	} else if rb, err = rb.Insert(15, []rune("5")); err != nil {
		t.Fatal(err)
	} else if err = validateRead(rb, []rune("1hell2o34 world5")); err != nil {
		t.Fatal(err)
		// after end
	} else if _, err = rb.Insert(17, []rune("a")); err != v.ErrorIndexOutofbound {
		t.Fatalf("expecting err ErrorIndexOutofbound")
	} else if err = validateRead(rb, []rune("1hell2o34 world5")); err != nil {
		t.Fatal(err)
	}
}

func TestRopeInsertLot(t *testing.T) {
	var err error

	rb := NewRopebuffer([]rune("hello world"), 2)
	lb := NewLinearBuffer([]rune("hello world"))
	lenrb := len(rb.Value())
	insVals := [][]rune{[]rune(""), []rune("a"), []rune("alpha")}
	for i := 0; i < 100; i++ {
		dot := int64(rand.Intn(lenrb))
		for _, insVal := range insVals {
			rb, err = rb.Insert(dot, insVal)
			if err != nil {
				t.Fatal(err)
			}
			lb, err = lb.Insert(dot, insVal)
			if err != nil {
				t.Fatal(err)
			}
			if err = validateRead(rb, nil); err != nil {
				t.Fatal(err)
			}
			if err = validateDicing(rb); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestRopeDeleteBasic(t *testing.T) {
	rb := NewRopebuffer([]rune("hello world"), 2)
	// before begin
	if _, err := rb.Delete(-1, 0); err != v.ErrorIndexOutofbound {
		t.Fatalf("expecting err ErrorIndexOutofbound")
	} else if err = validateRead(rb, []rune("hello world")); err != nil {
		t.Fatal(err)
		// at begin
	} else if rb, err = rb.Delete(0, 1); err != nil {
		t.Fatal(err)
	} else if err = validateRead(rb, []rune("ello world")); err != nil {
		t.Fatal(err)
		// before middle
	} else if rb, err = rb.Delete(1, 2); err != nil {
		t.Fatal(err)
	} else if err = validateRead(rb, []rune("eo world")); err != nil {
		t.Fatal(err)
		// after middle
	} else if rb, err = rb.Delete(2, 3); err != nil {
		t.Fatal(err)
	} else if err = validateRead(rb, []rune("eorld")); err != nil {
		t.Fatal(err)
		// at middle
	} else if _, err = rb.Delete(3, 4); err != v.ErrorIndexOutofbound {
		t.Fatalf("expecting err ErrorIndexOutofbound")
	} else if err = validateRead(rb, []rune("eorld")); err != nil {
		t.Fatal(err)
		// at end
	} else if rb, err = rb.Delete(2, 3); err != nil {
		t.Fatal(err)
	} else if err = validateRead(rb, []rune("eo")); err != nil {
		t.Fatal(err)
		// after end
	} else if _, err = rb.Delete(1, 1); err != nil {
		t.Fatal(err)
	} else if err = validateRead(rb, []rune("eo")); err != nil {
		t.Fatal(err)
	}
}

func TestRopePersistence(t *testing.T) {
	var err error
	rb := NewRopebuffer([]rune("hello world"), 2)
	lb := NewLinearBuffer(rb.Value())
	history := map[*RopeBuffer]bool{rb: true}
	for i := 0; i < 100; i++ {
		rb, err = rb.Insert(int64(i), []rune("abc"))
		if err != nil {
			t.Fatal(err)
		}
		lb, _ = lb.Insert(int64(i), []rune("abc"))
		if _, ok := history[rb]; ok {
			t.Fatal("persistence ref failed %v", rb.Value())
		}
		val := string(rb.Value())
		for oldrb := range history {
			if string(oldrb.Value()) == val {
				t.Fatal("persistence value failed %v", val)
			}
		}
		history[rb] = true
		if err := validateRead(rb, lb.Value()); err != nil {
			t.Fatal(err)
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
		rb.Insert(1024*512, runes)
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

func validateRead(rb *RopeBuffer, ref []rune) error {
	if ref != nil {
		// verify length
		if y, err := rb.Length(); err != nil {
			return err
		} else if int64(len(ref)) != y {
			return fmt.Errorf("expecting length %d, got %d", len(ref), y)
		}
		// verify value
		if x, y := string(rb.Value()), string(ref); x != y {
			return fmt.Errorf("expecting value %q, got %q", x, y)
		}
		// verify index
		for i, x := range ref {
			if y, ok, err := rb.Index(int64(i)); err != nil {
				return err
			} else if !ok {
				return fmt.Errorf("expecting rune at %d for %q", i, string(ref))
			} else if x != y {
				return fmt.Errorf("expecting %v, got %v at %d", x, y, i)
			}
		}
		// out of bound index
		if _, _, err := rb.Index(-1); err != v.ErrorIndexOutofbound {
			return fmt.Errorf("expecting v.ErrorIndexOutofbound at -1")
		}
		// out of bound index
		_, _, err := rb.Index(int64(len(ref)))
		if err != v.ErrorIndexOutofbound {
			return fmt.Errorf("expecting v.ErrorIndexOutofbound at %d", len(ref))
		}
		// verify substr
		for dot := 0; dot < len(ref); dot++ {
			for n := 0; n < len(ref)-dot; n++ {
				s, err := rb.Substr(int64(dot), int64(n))
				if err != nil {
					return err
				}
				if string(s) != string(ref[dot:dot+n]) {
					msg := "expecting %q, got %q"
					return fmt.Errorf(msg, string(s), string(ref[dot:n]))
				}
			}
		}
	}

	lb := NewLinearBuffer(rb.Value())
	ref = lb.Value()
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
		if err := validateRead(rb, nil); err != nil {
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
