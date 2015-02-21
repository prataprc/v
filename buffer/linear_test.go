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
		t.Fatalf("size expected %v, got %v\n", l, totalsize)
	} else if s := string(runes); s != testChinese {
		t.Fatalf("expected %v\ngot %v\n", testChinese, s)
	}
}

func TestLinearStreamCount(t *testing.T) {
	var err error
	var r rune
	var size int
	lb := NewLinearBuffer([]byte(testChinese))
	offs := runePositions([]byte(testChinese))
	reader := lb.StreamCount(0, 10)
	totalrn, totalsize, runes := 0, 0, make([]rune, 10)
	for err != io.EOF {
		r, size, err = reader.ReadRune()
		if err != io.EOF {
			runes[totalrn] = r
			totalrn++
			totalsize += size
		}
	}
	if totalrn != 10 {
		t.Fatalf("rune len expected %v, got %v\n", 10, totalrn)
	} else if int64(totalsize) != offs[10] {
		t.Fatalf("size expected %v, got %v\n", offs[10], totalsize)
	} else if s := string(runes); s != testChinese[:offs[10]] {
		t.Fatalf("expected %v\ngot %v\n", testChinese[:offs[10]], s)
	}
}

func TestLinearBackStreamFrom(t *testing.T) {
	var err error
	var r rune
	var size int
	lb := NewLinearBuffer([]byte(testChinese))
	offs := runePositions([]byte(testChinese))
	reader := lb.BackStreamFrom(offs[10])
	totalrn, totalsize, runes := 0, 0, make([]rune, 10)
	for err != io.EOF {
		r, size, err = reader.ReadRune()
		if err != io.EOF {
			runes[totalrn] = r
			totalrn++
			totalsize += size
		}
	}
	if totalrn != 10 {
		t.Fatalf("rune len expected %v, got %v\n", 10, totalrn)
	} else if int64(totalsize) != offs[10] {
		t.Fatalf("size expected %v, got %v\n", offs[10], totalsize)
	} else if s := string(reverseRunes(runes)); s != testChinese[:offs[10]] {
		t.Fatalf("expected %v\ngot %v\n", testChinese[:offs[10]], s)
	}
}

func TestLinearBackStreamCount(t *testing.T) {
	var err error
	var r rune
	var size int
	lb := NewLinearBuffer([]byte(testChinese))
	offs := runePositions([]byte(testChinese))
	reader := lb.BackStreamCount(offs[10], 9)
	totalrn, totalsize, runes := 0, 0, make([]rune, 9)
	for err != io.EOF {
		r, size, err = reader.ReadRune()
		if err != io.EOF {
			runes[totalrn] = r
			totalrn++
			totalsize += size
		}
	}
	s := string(reverseRunes(runes))
	if totalrn != 9 {
		t.Fatalf("rune len expected %v, got %v\n", 9, totalrn)
	} else if l := offs[10] - offs[1]; int64(totalsize) != l {
		t.Fatalf("size expected %v, got %v\n", totalsize, l)
	} else if ref := testChinese[offs[1]:offs[10]]; s != ref {
		t.Fatalf("expected %v\ngot %v\n", ref, s)
	}
}

func BenchmarkLinearLength(b *testing.B) {
	lb := NewLinearBuffer([]byte(testChinese))
	for i := 0; i < b.N; i++ {
		if _, err := lb.Length(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLinearValue(b *testing.B) {
	lb := NewLinearBuffer([]byte(testChinese))
	for i := 0; i < b.N; i++ {
		lb.Value()
	}
}

func BenchmarkLinearSlice(b *testing.B) {
	lb := NewLinearBuffer([]byte(testChinese))
	l, err := lb.Length()
	if err != nil {
		b.Fatal(err)
	}
	size := 0
	for i := 0; i < b.N; i++ {
		bs, err := lb.Slice(0, int64(i)%l)
		if err != nil {
			b.Fatal(err)
		}
		size += len(bs)
	}
	b.SetBytes(int64(size / b.N))
}

func BenchmarkLinearRuneAt(b *testing.B) {
	lb := NewLinearBuffer([]byte(testChinese))
	offs := runePositions([]byte(testChinese))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, _, err := lb.RuneAt(offs[i%len(offs)]); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLinearRunes(b *testing.B) {
	lb := NewLinearBuffer([]byte(testChinese))
	for i := 0; i < b.N; i++ {
		if _, err := lb.Runes(); err != nil {
			b.Fatal(err)
		}
	}
	b.SetBytes(int64(len(testChinese)))
}

func BenchmarkLinearRuneSl(b *testing.B) {
	lb := NewLinearBuffer([]byte(testChinese))
	offs := runePositions([]byte(testChinese))
	b.ResetTimer()
	size := int64(0)
	for i := 0; i < b.N; i++ {
		_, sz, err := lb.RuneSlice(0, offs[i%len(offs)])
		if err != nil {
			b.Fatal(err)
		}
		size += sz
	}
	b.SetBytes(size / int64(b.N))
}

func BenchmarkLinearConcat(b *testing.B) {
	lb1 := NewLinearBuffer([]byte(testChinese))
	lb2 := NewLinearBuffer([]byte(testChinese))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := lb1.Concat(lb2); err != nil {
			b.Fatal(err)
		}
	}
	b.SetBytes(int64(len(testChinese) * 2))
}

func BenchmarkLinearSplit(b *testing.B) {
	lb := NewLinearBuffer([]byte(testChinese))
	l, _ := lb.Length()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, _, err := lb.Split(int64(i) % l); err != nil {
			b.Fatal(err)
		}
	}
	b.SetBytes(int64(len(testChinese)))
}

func BenchmarkLinearInsert(b *testing.B) {
	// insert small text into a small buffer.
	lb := NewLinearBuffer([]byte(testChinese))
	itext := []rune(`中國;pinyin`)
	offs := runePositions([]byte(testChinese))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bCur := offs[i%len(offs)]
		if _, err := lb.Insert(bCur, itext); err != nil {
			b.Fatal(err)
		}
	}
	b.SetBytes(int64(len(string(itext))))
}

func BenchmarkLinearInsert2M(b *testing.B) {
	// insert small text into a large buffer.
	lb := NewLinearBuffer([]byte(sampleData))
	itext := []rune(`中國;pinyin`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bCur := i % len(sampleData)
		if _, err := lb.Insert(int64(bCur), itext); err != nil {
			b.Fatal(err)
		}
	}
	b.SetBytes(int64(len(string(itext))))
}

func BenchmarkLinearDelete(b *testing.B) {
	// insert small text from a small buffer.
	lb := NewLinearBuffer([]byte(testChinese))
	itext := []rune(`中國;pinyin`)
	offs := runePositions([]byte(testChinese))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bCur := offs[:10][int64(i)%10]
		if _, err := lb.Delete(bCur, 10); err != nil {
			b.Fatal(err)
		}
		if _, err := lb.Insert(bCur, itext); err != nil {
			b.Fatal(err)
		}
	}
	b.SetBytes(int64(len(string(itext))))
}

func BenchmarkLinearDelete2M(b *testing.B) {
	// insert small text from a small buffer.
	lb := NewLinearBuffer([]byte(sampleData))
	itext := []rune(`中國;pinyin`)
	offs := runePositions([]byte(sampleData))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bCur := offs[:1000][int64(i)%1000]
		if _, err := lb.Delete(bCur, 10); err != nil {
			b.Fatal(err)
		}
		if _, err := lb.Insert(bCur, itext); err != nil {
			b.Fatal(err)
		}
	}
	b.SetBytes(int64(len(string(itext))))
}

func BenchmarkLinearInsIn(b *testing.B) {
	// insert small text into a small buffer.
	itext := []rune(`中國;pinyin`)
	offs := runePositions([]byte(testChinese))
	bCur := offs[10]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lb := NewLinearBuffer([]byte(testChinese))
		if _, err := lb.InsertIn(bCur, itext); err != nil {
			b.Fatal(err)
		}
	}
	b.SetBytes(int64(len(string(itext))))
}

func BenchmarkLinearInsIn2M(b *testing.B) {
	// insert small text into a large buffer.
	itext := []rune(`中國;pinyin`)
	offs := runePositions([]byte(sampleData))
	bCur := offs[1000]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lb := NewLinearBuffer([]byte(sampleData))
		if _, err := lb.InsertIn(bCur, itext); err != nil {
			b.Fatal(err)
		}
	}
	b.SetBytes(int64(len(string(itext))))
}

func BenchmarkLinearDelIn(b *testing.B) {
	// insert small text from a small buffer.
	lb := NewLinearBuffer([]byte(testChinese))
	itext := []rune(`中國;pinyin`)
	offs := runePositions([]byte(testChinese))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bCur := offs[:10][int64(i)%10]
		if _, err := lb.DeleteIn(bCur, 10); err != nil {
			b.Fatal(err)
		}
		if _, err := lb.InsertIn(bCur, itext); err != nil {
			b.Fatal(err)
		}
	}
	b.SetBytes(int64(len(string(itext))))
}

func BenchmarkLinearDelIn2M(b *testing.B) {
	// insert small text from a small buffer.
	lb := NewLinearBuffer([]byte(sampleData))
	itext := []rune(`中國;pinyin`)
	offs := runePositions([]byte(sampleData))
	bCur := offs[1000]
	lb.InsertIn(bCur, itext)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := lb.DeleteIn(bCur, 10); err != nil {
			b.Fatal(err)
		}
		if _, err := lb.InsertIn(bCur, itext); err != nil {
			b.Fatal(err)
		}
	}
	b.SetBytes(int64(len(string(itext))))
}

func BenchmarkLinearStrmFrm(b *testing.B) {
	lb := NewLinearBuffer([]byte(testChinese))
	for i := 0; i < b.N; i++ {
		var err error
		reader := lb.StreamFrom(0)
		for err != io.EOF {
			_, _, err = reader.ReadRune()
		}
	}
	b.SetBytes(int64(len(testChinese)))
}

func BenchmarkLinearStrmCnt(b *testing.B) {
	lb := NewLinearBuffer([]byte(testChinese))
	offs := runePositions([]byte(testChinese))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var err error
		reader := lb.StreamCount(0, int64(len(offs)))
		for err != io.EOF {
			_, _, err = reader.ReadRune()
		}
	}
	b.SetBytes(int64(len(testChinese)))
}

func BenchmarkLinearBStrmFrm(b *testing.B) {
	lb := NewLinearBuffer([]byte(testChinese))
	offs := runePositions([]byte(testChinese))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var err error
		reader := lb.BackStreamFrom(offs[len(offs)-1])
		for err != io.EOF {
			_, _, err = reader.ReadRune()
		}
	}
	b.SetBytes(int64(len(testChinese)))
}

func BenchmarkLinearBStrmCnt(b *testing.B) {
	lb := NewLinearBuffer([]byte(testChinese))
	offs := runePositions([]byte(testChinese))
	for i := 0; i < b.N; i++ {
		var err error
		reader := lb.BackStreamCount(offs[len(offs)-1], int64(len(offs)))
		for err != io.EOF {
			_, _, err = reader.ReadRune()
		}
	}
	b.SetBytes(int64(len(testChinese)))
}
