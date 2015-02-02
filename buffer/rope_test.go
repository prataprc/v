package buffer

import "testing"
import "io"
import "fmt"
import "log"
import "math/rand"
import "io/ioutil"

var testRopeBufferCapacity = int64(10 * 1024) // 10KB nodes in rope.
var sampleData []byte                         // contains 2.5MB data.

func init() {
	var err error
	sampleData, err = ioutil.ReadFile("../tools/monstrun/sample.txt")
	if err != nil {
		log.Fatal(err)
	}
}

func TestRopeSample256(t *testing.T) {
	rb, err := NewRopebuffer(sampleData, 256)
	if err != nil {
		t.Fatal(err)
	}
	stats := rb.Stats()
	length := validateRopeBuild(t, stats)
	if l := int64(len(sampleData)); length != l {
		t.Fatalf("mismatch in length %v, got %v", length, l)
	}
}

func TestRopeSample1MB(t *testing.T) {
	rb, err := NewRopebuffer(sampleData, 1024*1024)
	if err != nil {
		t.Fatal(err)
	}
	stats := rb.Stats()
	length := validateRopeBuild(t, stats)
	if l := int64(len(sampleData)); length != l {
		t.Fatalf("mismatch in length %v, got %v", length, l)
	}
}

func TestRopeIndex(t *testing.T) {
	rb, err := NewRopebuffer([]byte("hello world"), 2)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateRead(rb, nil); err != nil {
		t.Fatal(err)
	}
}

func TestRopeRuneSlice(t *testing.T) {
	rb, err := NewRopebuffer([]byte("hello world"), 2)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateRuneSlice(rb); err != nil {
		t.Fatal(err)
	}
}

func TestRopeDicing(t *testing.T) {
	rb, err := NewRopebuffer([]byte("hello world"), 2)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateDicing(rb); err != nil {
		t.Fatal(err)
	}
}

func TestRopeInsertBasic(t *testing.T) {
	rb, err := NewRopebuffer([]byte("hello world"), 2)
	if err != nil {
		t.Fatal(err)
	}
	// before begin
	if _, err := rb.Insert(-1, []rune("a")); err != ErrorIndexOutofbound {
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
	} else if _, err = rb.Insert(17, []rune("a")); err != ErrorIndexOutofbound {
		t.Fatalf("expecting err ErrorIndexOutofbound")
	} else if err = validateRead(rb, []rune("1hell2o34 world5")); err != nil {
		t.Fatal(err)
	}
}

func TestRopeInsertMany(t *testing.T) {
	var err error

	rb, err := NewRopebuffer([]byte("hello world"), 2)
	if err != nil {
		t.Fatal(err)
	}
	lb := NewLinearBuffer([]byte("hello world"))
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
	rb, err := NewRopebuffer([]byte("hello world"), 2)
	if err != nil {
		t.Fatal(err)
	}
	// before begin
	if _, err := rb.Delete(-1, 0); err != nil {
		t.Fatalf("unexpected err")
	} else if _, err := rb.Delete(-1, 1); err != ErrorIndexOutofbound {
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
	} else if _, err = rb.Delete(3, 4); err != ErrorIndexOutofbound {
		t.Log(string(rb.Value()))
		t.Fatalf("expecting err ErrorIndexOutofbound, got %v", err)
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
	rb, err := NewRopebuffer([]byte("hello world"), 2)
	if err != nil {
		t.Fatal(err)
	}
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
		ref, err := lb.Runes()
		if err != nil {
			t.Fatal(err)
		}
		if err := validateRead(rb, ref); err != nil {
			t.Fatal(err)
		}
	}
}

func TestRopeStreamFrom(t *testing.T) {
	rb, err := NewRopebuffer([]byte("hello world"), 2)
	if err != nil {
		t.Fatal(err)
	}
	for i := -1; int64(i) < rb.Len+1; i++ {
		runes := make([]rune, 0)
		reader := rb.StreamFrom(int64(i))
		r, _, err := reader.ReadRune()
		for err != io.EOF {
			runes = append(runes, r)
			r, _, err = reader.ReadRune()
		}
		if i == -1 || int64(i) == rb.Len {
			if len(runes) != 0 {
				t.Fatalf("mismatch for %d %q", i, string(runes))
			}
		} else if x, y := string(runes), string(rb.Value()[i:]); x != y {
			t.Fatalf("mismatch for %d %q %q", i, x, y)
		}
	}
}

func TestRopeStreamTill(t *testing.T) {
	rb, err := NewRopebuffer([]byte("hello world"), 2)
	if err != nil {
		t.Fatal(err)
	}
	for i := -1; int64(i) < 6; i++ {
		runes := make([]rune, 0)
		reader := rb.StreamTill(int64(i), 6)
		r, _, err := reader.ReadRune()
		for err != io.EOF {
			runes = append(runes, r)
			r, _, err = reader.ReadRune()
		}
		if i == -1 || int64(i) == rb.Len {
			if len(runes) != 0 {
				t.Fatalf("mismatch for %d %q", i, string(runes))
			}
		} else if x, y := string(runes), string(rb.Value()[i:6]); x != y {
			t.Fatalf("mismatch for %d %q %q", i, x, y)
		}
	}
}

func BenchmarkRopeSample8(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewRopebuffer(sampleData, 8)
	}
	b.SetBytes(int64(len(sampleData)))
}

func BenchmarkRopeSample256(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewRopebuffer(sampleData, 256)
	}
	b.SetBytes(int64(len(sampleData)))
}

func BenchmarkRopeLength(b *testing.B) {
	rb, err := NewRopebuffer(sampleData, testRopeBufferCapacity)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb.Length()
	}
}

func BenchmarkRopeValue(b *testing.B) {
	rb, err := NewRopebuffer(sampleData, testRopeBufferCapacity)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb.Value()
	}
}

func BenchmarkRopeRuneAt(b *testing.B) {
	rb, err := NewRopebuffer(sampleData, testRopeBufferCapacity)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb.RuneAt(1024 * 512)
	}
}

func BenchmarkRopeRuneSlice(b *testing.B) {
	rb, err := NewRopebuffer(sampleData, testRopeBufferCapacity)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb.RuneSlice(1024*512, 10000)
	}
}

func BenchmarkRopeInsert1(b *testing.B) {
	rb, err := NewRopebuffer(sampleData, testRopeBufferCapacity)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb.Insert(104*512, []rune{'a'})
	}
}

func BenchmarkRopeInsert100(b *testing.B) {
	rb, err := NewRopebuffer(sampleData, testRopeBufferCapacity)
	if err != nil {
		b.Fatal(err)
	}
	runes, _, err := rb.RuneSlice(0, 100)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb.Insert(104*512, runes)
	}
}

func BenchmarkRopeInsertIn1(b *testing.B) {
	rb, err := NewRopebuffer(sampleData, testRopeBufferCapacity)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb.InsertIn(104*512, []rune{'a'})
	}
}

func BenchmarkRopeInsertIn100(b *testing.B) {
	rb, err := NewRopebuffer(sampleData, testRopeBufferCapacity)
	if err != nil {
		b.Fatal(err)
	}
	runes, _, err := rb.RuneSlice(0, 100)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb.InsertIn(104*512, runes)
	}
}

func BenchmarkRopeDelete1(b *testing.B) {
	rb, err := NewRopebuffer(sampleData, testRopeBufferCapacity)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb.Delete(104*512, 1)
	}
}

func BenchmarkRopeDelete1000(b *testing.B) {
	rb, err := NewRopebuffer(sampleData, testRopeBufferCapacity)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb.Delete(104*512, 1000)
	}
}

func BenchmarkRopeDeleteIn1(b *testing.B) {
	rb, err := NewRopebuffer(sampleData, testRopeBufferCapacity)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb.DeleteIn(104*512, 1)
	}
}

func BenchmarkRopeDeleteIn1000(b *testing.B) {
	rb, err := NewRopebuffer(sampleData, testRopeBufferCapacity)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb.DeleteIn(104*512, 1000)
	}
}

func BenchmarkBytes2Str(b *testing.B) {
	var text string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		text = string(sampleData)
	}
	b.SetBytes(int64(len(text)))
}

func BenchmarkStr2Byte(b *testing.B) {
	var bs []byte
	text := string(sampleData)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bs = []byte(text)
	}
	b.SetBytes(int64(len(bs)))
}

func BenchmarkStr2Runes(b *testing.B) {
	var runes []rune
	str := string(sampleData)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runes = []rune(str)
	}
	b.SetBytes(int64(len(runes)))
}

func BenchmarkRunes2Str(b *testing.B) {
	var str string
	runes := []rune(string(sampleData))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		str = string(runes)
	}
	b.SetBytes(int64(len(str)))
}

func BenchmarkStreamFrom(b *testing.B) {
	rb, err := NewRopebuffer(sampleData, testRopeBufferCapacity)
	if err != nil {
		b.Fatal(err)
	}
	count := 0
	for i := 0; i < b.N; i++ {
		reader := rb.StreamFrom(0)
		_, _, err := reader.ReadRune()
		for err != io.EOF {
			_, _, err = reader.ReadRune()
			count++
		}
	}
	b.SetBytes(int64(len(sampleData)))
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
	if ref := 1 << (uint64(maxLevel) - 1); int64(ref) != leafs {
		t.Fatalf("unexpected leafs %v, got %v", ref, leafs)
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
			if y, size, err := rb.RuneAt(int64(i)); err != nil {
				return err
			} else if size == 0 {
				return fmt.Errorf("expecting rune at %d for %q", i, string(ref))
			} else if x != y {
				return fmt.Errorf("expecting %v, got %v at %d", x, y, i)
			}
		}
		// out of bound index
		if _, _, err := rb.RuneAt(-1); err != ErrorIndexOutofbound {
			return fmt.Errorf("expecting ErrorIndexOutofbound at -1")
		}
		// out of bound index
		_, _, err := rb.RuneAt(int64(len(ref)))
		if err != ErrorIndexOutofbound {
			return fmt.Errorf("expecting ErrorIndexOutofbound at %d", len(ref))
		}
		// verify substr
		for dot := 0; dot < len(ref); dot++ {
			for n := 0; n < len(ref)-dot; n++ {
				runes1, _, err := rb.RuneSlice(int64(dot), int64(n))
				if err != nil {
					return err
				}
				if string(runes1) != string(ref[dot:dot+n]) {
					msg := "expecting %q, got %q"
					return fmt.Errorf(msg, string(runes1), string(ref[dot:n]))
				}
			}
		}
	}

	lb := NewLinearBuffer(rb.Value())
	ref, err := lb.Runes()
	if err != nil {
		return err
	}
	rbRunes, err := rb.Runes()
	if err != nil {
		return err
	}

	if x, y := string(rbRunes), string(ref); x != y {
		return fmt.Errorf("mismatch in value %q, got %q", x, y)
	}

	length, err := rb.Length()
	if err != nil {
		return err
	}
	if length == 0 {
		if _, size, err := rb.RuneAt(0); err == nil || size > 0 {
			return fmt.Errorf("expecting error")
		}
		if _, size, err := rb.RuneAt(1); err == nil || size > 0 {
			return fmt.Errorf("expecting error")
		}
	} else {
		for _, dot := range []int64{0, length / 2, length - 1} {
			if ch, size, err := rb.RuneAt(dot); err != nil {
				return err
			} else if size == 0 {
				return fmt.Errorf("decoded 0 bytes at %v ch %v", dot, ref[dot])
			} else if ch != ref[dot] {
				return fmt.Errorf("mismatch for ch %v, got %v", ref[dot], ch)
			}
		}
		if _, size, err := rb.RuneAt(length); err == nil || size > 0 {
			return fmt.Errorf("expecting error for %v", length)
		}
	}

	return validateRuneSlice(rb)
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

func validateRuneSlice(rbRef *RopeBuffer) error {
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
		rbRunes, _, err := rbRef.RuneSlice(dot, n)
		if err != nil {
			return fmt.Errorf("rb at %d size %d for %q", dot, n, string(x))
		}
		lbRunes, _, err := lb.RuneSlice(dot, n)
		if err != nil {
			return fmt.Errorf("lb at %d size %d for %q", dot, n, string(x))
		}
		if string(rbRunes) != string(lbRunes) {
			return fmt.Errorf("at %d size %d for %q", dot, n, string(x))
		}
	}
	return nil
}
