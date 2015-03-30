// +build ignore

package buffer

import "testing"
import "reflect"
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
    stats, err := rb.Stats()
    if err != nil {
        t.Fatal(err)
    }
    length := validateRopeBuild(t, stats)
    if l := int64(len(sampleData)); length != l {
        t.Fatalf("mismatch in length %v, got %v", length, l)
    } else if v := string(rb.Value()); string(sampleData) != v {
        t.Fatalf("mismatch expected %v, got %v", len(sampleData), len(v))
    }
}

func TestRopeSample1MB(t *testing.T) {
    rb, err := NewRopebuffer(sampleData, 1024*1024)
    if err != nil {
        t.Fatal(err)
    }
    stats, err := rb.Stats()
    if err != nil {
        t.Fatal(err)
    }
    length := validateRopeBuild(t, stats)
    if l := int64(len(sampleData)); length != l {
        t.Fatalf("mismatch in length %v, got %v", length, l)
    }
}

func TestRopeIndex(t *testing.T) {
    rb, err := NewRopebuffer([]byte(testChinese), 8)
    if err != nil {
        t.Fatal(err)
    }
    runes := []rune(testChinese)
    if err := validateRead(rb, runes); err != nil {
        t.Fatal(err)
    }
}

func TestRopeRuneSlice(t *testing.T) {
    rb, err := NewRopebuffer([]byte(testChinese), 8)
    if err != nil {
        t.Fatal(err)
    }
    if err := validateRuneSlice(rb); err != nil {
        t.Fatal(err)
    }
}

func TestRopeDicing(t *testing.T) {
    rb, err := NewRopebuffer([]byte(testChinese), 8)
    if err != nil {
        t.Fatal(err)
    }
    if err := validateDicing(rb); err != nil {
        t.Fatal(err)
    }
}

func TestRopeInsert(t *testing.T) {
    rb, err := NewRopebuffer([]byte(testChinese), 8)
    if err != nil {
        t.Fatal(err)
    }
    if _, err := rb.Insert(-1, []rune("a")); err != ErrorIndexOutofbound {
        t.Fatalf("expecting err ErrorIndexOutofbound")
    } else if _, err = rb.Insert(170, []rune("a")); err != ErrorIndexOutofbound {
        t.Fatalf("expecting err ErrorIndexOutofbound")
    }
    offs := runePositions([]byte(testChinese))
    for _, dot := range offs {
        rb, err := rb.Insert(dot, []rune("道"))
        if err != nil {
            t.Fatal(err)
        }
        runes := []rune(string(rb.Value()))
        if err := validateRead(rb.(*RopeBuffer), runes); err != nil {
            t.Fatal(err)
        }
    }
}

func TestRopeDelete(t *testing.T) {
    rb, err := NewRopebuffer([]byte(testChinese), 8)
    if err != nil {
        t.Fatal(err)
    }
    var buf Buffer
    for rb.Len > 0 {
        offs := runePositions(rb.Value())
        i := rand.Intn(len(offs))
        n := rand.Intn(len(offs) - i + 1)
        dot := offs[i]
        if buf, err = rb.Delete(dot, int64(n)); err != nil {
            t.Fatal(err)
        }
        rb = buf.(*RopeBuffer)
    }
}

func TestRopePersistence(t *testing.T) {
    offs := runePositions([]byte(testChinese))
    rb, err := NewRopebuffer([]byte(testChinese), 8)
    if err != nil {
        t.Fatal(err)
    }
    lb := NewLinearBuffer(rb.Value())
    history := map[*RopeBuffer]bool{rb: true}
    for _, off := range offs {
        buf, err := rb.Insert(off, []rune("abc"))
        if err != nil {
            t.Fatal(err)
        }
        rb = buf.(*RopeBuffer)
        lb, _ = lb.Insert(off, []rune("abc"))
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

func TestJohnnie(t *testing.T) {
    rb, err := NewRopebuffer([]byte(testChinese), 8)
    if err != nil {
        t.Fatal(err)
    }
    rbs1 := make(map[*RopeBuffer]bool)
    rb.Walk(0, func(bCur int64, rb *RopeBuffer) bool {
        rbs1[rb] = true
        return true
    })

    rbs2 := make(map[*RopeBuffer]bool)
    l, err := rb.Length()
    if err != nil {
        t.Fatal(err)
    }
    rb.WalkBack(l, func(bCur int64, rb *RopeBuffer) bool {
        rbs2[rb] = true
        return true
    })
    if !reflect.DeepEqual(rbs1, rbs2) {
        t.Fatalf("mismatch in leaf nodes")
    }

    runes1 := make([]rune, 0)
    for reader := rb.StreamFrom(0); true; {
        r, _, err := reader.ReadRune()
        if err == io.EOF {
            break
        }
        runes1 = append(runes1, r)
    }
    runes2 := make([]rune, 0)
    for reader := rb.BackStreamFrom(l); true; {
        r, _, err := reader.ReadRune()
        if err == io.EOF {
            break
        }
        runes2 = append(runes2, r)
    }
    runes2 = reverseRunes(runes2)
    if x, y := string(runes1), string(runes2); x != y {
        t.Fatalf("expected %s, got %s\n", x, y)
    }
}

func TestRopeStreamFrom(t *testing.T) {
    bytes := []byte(testChinese)
    rb, err := NewRopebuffer(bytes, 8)
    if err != nil {
        t.Fatal(err)
    }
    offs := runePositions(bytes)
    offs = append(append([]int64{}, -1), offs...)
    for _, off := range offs {
        runes := make([]rune, 0)
        for reader := rb.StreamFrom(off); true; {
            r, _, err := reader.ReadRune()
            if err == io.EOF {
                break
            }
            runes = append(runes, r)
        }
        if off == -1 || off == rb.Len {
            if len(runes) != 0 {
                t.Fatalf("mismatch for %d %q", off, string(runes))
            }
        } else if x, y := string(runes), string(rb.Value()[off:]); x != y {
            t.Fatalf("mismatch for %d %q, got %q", off, y, x)
        }
    }
}

func TestRopeStreamCount(t *testing.T) {
    bytes := []byte(testChinese)
    rb, err := NewRopebuffer(bytes, 8)
    if err != nil {
        t.Fatal(err)
    }
    ref, err := rb.Runes()
    if err != nil {
        t.Fatal(err)
    }
    offs := runePositions(bytes)
    for i, off := range offs {
        runes := make([]rune, 0)
        n := rand.Intn(len(offs) - i)
        for reader := rb.StreamCount(off, int64(n)); true; {
            r, _, err := reader.ReadRune()
            if err == io.EOF {
                break
            }
            runes = append(runes, r)
        }
        x, y := string(runes), string(ref[i:i+n])
        if n == i {
            if len(runes) != 0 {
                t.Fatalf("mismatch for %d %q", off, string(runes))
            }
        } else if x != y {
            t.Fatalf("mismatch for %d %q, got %q", off, y, x)
        }
    }
}

func TestRopeBackStreamFrom(t *testing.T) {
    bytes := []byte(testChinese)
    rb, err := NewRopebuffer(bytes, 8)
    if err != nil {
        t.Fatal(err)
    }
    ref, err := rb.Runes()
    if err != nil {
        t.Fatal(err)
    }
    offs := runePositions(bytes)
    for i, off := range offs {
        runes := make([]rune, 0)
        for reader := rb.BackStreamFrom(off); true; {
            r, _, err := reader.ReadRune()
            if err == io.EOF {
                break
            }
            runes = append(runes, r)
        }
        runes = reverseRunes(runes)
        if off == rb.Len {
            if len(runes) != 0 {
                t.Fatalf("mismatch for %d %q", off, string(runes))
            }
        } else if x, y := string(runes), string(ref[0:i+1]); x != y {
            t.Fatalf("mismatch for {%d,%d} %q, got %q", off, i, y, x)
        }
    }
}

func TestRopeBackStreamCount(t *testing.T) {
    bytes := []byte(testChinese)
    rb, err := NewRopebuffer(bytes, 8)
    if err != nil {
        t.Fatal(err)
    }
    ref, err := rb.Runes()
    if err != nil {
        t.Fatal(err)
    }
    offs := runePositions(bytes)
    for i, off := range offs {
        runes := make([]rune, 0)
        n := rand.Intn(i + 1)
        for reader := rb.BackStreamCount(off, int64(n)); true; {
            r, _, err := reader.ReadRune()
            if err == io.EOF {
                break
            }
            runes = append(runes, r)
        }
        runes = reverseRunes(runes)
        if x, y := string(runes), string(ref[i-n+1:i+1]); x != y {
            t.Fatalf("mismatch for {%d,%d,%d} %q, got %q", off, n, i, y, x)
        }
    }
}

func BenchmarkRopeSample8(b *testing.B) {
    bytes := []byte(testChinese)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        NewRopebuffer(bytes, 8)
    }
    b.SetBytes(int64(len(testChinese)))
}

func BenchmarkRopeLength(b *testing.B) {
    rb, err := NewRopebuffer([]byte(testChinese), 8)
    if err != nil {
        b.Fatal(err)
    }
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        if _, err := rb.Length(); err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkRopeValue(b *testing.B) {
    rb, err := NewRopebuffer([]byte(testChinese), 8)
    if err != nil {
        b.Fatal(err)
    }
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        rb.Value()
    }
}

func BenchmarkRopeSlice(b *testing.B) {
    rb, err := NewRopebuffer([]byte(testChinese), 8)
    if err != nil {
        b.Fatal(err)
    }
    l, err := rb.Length()
    if err != nil {
        b.Fatal(err)
    }
    size := 0
    for i := 0; i < b.N; i++ {
        bs, err := rb.Slice(0, int64(i)%l)
        if err != nil {
            b.Fatal(err)
        }
        size += len(bs)
    }
    b.SetBytes(int64(size / b.N))
}

func BenchmarkRopeRuneAt(b *testing.B) {
    rb, err := NewRopebuffer([]byte(testChinese), 8)
    if err != nil {
        b.Fatal(err)
    }
    offs := runePositions([]byte(testChinese))
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        if _, _, err := rb.RuneAt(offs[i%len(offs)]); err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkRopeRunes(b *testing.B) {
    rb, err := NewRopebuffer([]byte(testChinese), 8)
    if err != nil {
        b.Fatal(err)
    }
    for i := 0; i < b.N; i++ {
        if _, err := rb.Runes(); err != nil {
            b.Fatal(err)
        }
    }
    b.SetBytes(int64(len(testChinese)))
}

func BenchmarkRopeRuneSlice(b *testing.B) {
    rb, err := NewRopebuffer([]byte(testChinese), 8)
    if err != nil {
        b.Fatal(err)
    }
    offs := runePositions([]byte(testChinese))
    b.ResetTimer()
    size := int64(0)
    for i := 0; i < b.N; i++ {
        _, sz, err := rb.RuneSlice(0, offs[i%len(offs)])
        if err != nil {
            b.Fatal(err)
        }
        size += sz
    }
    b.SetBytes(size / int64(b.N))
}

func BenchmarkRopeConcat(b *testing.B) {
    rb1, _ := NewRopebuffer([]byte(testChinese), 8)
    rb2, _ := NewRopebuffer([]byte(testChinese), 8)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        if _, err := rb1.Concat(rb2); err != nil {
            b.Fatal(err)
        }
    }
    b.SetBytes(int64(len(testChinese) * 2))
}

func BenchmarkRopeSplit(b *testing.B) {
    rb, _ := NewRopebuffer([]byte(testChinese), 8)
    l, _ := rb.Length()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        if _, _, err := rb.Split(int64(i) % l); err != nil {
            b.Fatal(err)
        }
    }
    b.SetBytes(int64(len(testChinese)))
}

func BenchmarkRopeInsert(b *testing.B) {
    // insert small text into a small buffer.
    rb, _ := NewRopebuffer([]byte(testChinese), 8)
    itext := []rune(`中國;pinyin`)
    offs := runePositions([]byte(testChinese))
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        bCur := offs[i%len(offs)]
        if _, err := rb.Insert(bCur, itext); err != nil {
            b.Fatal(err)
        }
    }
    b.SetBytes(int64(len(string(itext))))
}

func BenchmarkRopeInsert2M(b *testing.B) {
    // insert small text into a large buffer.
    rb, _ := NewRopebuffer([]byte(sampleData), 256)
    itext := []rune(`中國;pinyin`)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        bCur := i % len(sampleData)
        if _, err := rb.Insert(int64(bCur), itext); err != nil {
            b.Fatal(err)
        }
    }
    b.SetBytes(int64(len(string(itext))))
}

func BenchmarkRopeDelete(b *testing.B) {
    rb, err := NewRopebuffer([]byte(testChinese), 8)
    if err != nil {
        b.Fatal(err)
    }
    itext := []rune(`中國;pinyin`)
    offs := runePositions([]byte(testChinese))
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        bCur := offs[:10][int64(i)%10]
        if _, err := rb.Delete(bCur, 10); err != nil {
            b.Fatal(err)
        }
        if _, err := rb.Insert(bCur, itext); err != nil {
            b.Fatal(err)
        }
    }
    b.SetBytes(int64(len(string(itext))))
}

func BenchmarkRopeDelete2M(b *testing.B) {
    // insert small text from a small buffer.
    rb, _ := NewRopebuffer([]byte(sampleData), 256)
    itext := []rune(`中國;pinyin`)
    offs := runePositions([]byte(sampleData))
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        bCur := offs[:1000][int64(i)%1000]
        if _, err := rb.Delete(bCur, 10); err != nil {
            b.Fatal(err)
        }
        if _, err := rb.Insert(bCur, itext); err != nil {
            b.Fatal(err)
        }
    }
    b.SetBytes(int64(len(string(itext))))
}

func BenchmarkRopeInsIn(b *testing.B) {
    // insert small text into a small buffer.
    itext := []rune(`中國;pinyin`)
    offs := runePositions([]byte(testChinese))
    bCur := offs[10]
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        rb, _ := NewRopebuffer([]byte(testChinese), 8)
        if _, err := rb.InsertIn(bCur, itext); err != nil {
            b.Fatal(err)
        }
    }
    b.SetBytes(int64(len(string(itext))))
}

func BenchmarkRopeInsIn2M(b *testing.B) {
    // insert small text into a large buffer.
    itext := []rune(`中國;pinyin`)
    offs := runePositions([]byte(sampleData))
    bCur := offs[1000]
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        rb, _ := NewRopebuffer([]byte(sampleData), 256)
        if _, err := rb.InsertIn(bCur, itext); err != nil {
            b.Fatal(err)
        }
    }
    b.SetBytes(int64(len(string(itext))))
}

func BenchmarkRopeDelIn(b *testing.B) {
    // insert small text from a small buffer.
    rb, _ := NewRopebuffer([]byte(testChinese), 8)
    itext := []rune(`中國;pinyin`)
    offs := runePositions([]byte(testChinese))
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        bCur := offs[:10][int64(i)%10]
        if _, err := rb.DeleteIn(bCur, 10); err != nil {
            b.Fatal(err)
        }
        if _, err := rb.InsertIn(bCur, itext); err != nil {
            b.Fatal(err)
        }
    }
    b.SetBytes(int64(len(string(itext))))
}

func BenchmarkRopeDelIn2M(b *testing.B) {
    // insert small text from a small buffer.
    rb, _ := NewRopebuffer([]byte(sampleData), 256)
    itext := []rune(`中國;pinyin`)
    offs := runePositions([]byte(sampleData))
    bCur := offs[1000]
    rb.InsertIn(bCur, itext)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        if _, err := rb.DeleteIn(bCur, 10); err != nil {
            b.Fatal(err)
        }
        if _, err := rb.InsertIn(bCur, itext); err != nil {
            b.Fatal(err)
        }
    }
    b.SetBytes(int64(len(string(itext))))
}

func BenchmarkRopeStrmFrm(b *testing.B) {
    rb, _ := NewRopebuffer([]byte(testChinese), 8)
    for i := 0; i < b.N; i++ {
        var err error
        reader := rb.StreamFrom(0)
        for err != io.EOF {
            _, _, err = reader.ReadRune()
        }
    }
    b.SetBytes(int64(len(testChinese)))
}

func BenchmarkRopeStrmCnt(b *testing.B) {
    rb, _ := NewRopebuffer([]byte(testChinese), 8)
    offs := runePositions([]byte(testChinese))
    for i := 0; i < b.N; i++ {
        var err error
        reader := rb.StreamCount(0, int64(len(offs)))
        for err != io.EOF {
            _, _, err = reader.ReadRune()
        }
    }
    b.SetBytes(int64(len(testChinese)))
}

func BenchmarkRopeBStrmFrm(b *testing.B) {
    rb, _ := NewRopebuffer([]byte(testChinese), 8)
    offs := runePositions([]byte(testChinese))
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        var err error
        reader := rb.BackStreamFrom(offs[len(offs)-1])
        for err != io.EOF {
            _, _, err = reader.ReadRune()
        }
    }
    b.SetBytes(int64(len(testChinese)))
}

func BenchmarkRopeBStrmCnt(b *testing.B) {
    rb, _ := NewRopebuffer([]byte(testChinese), 8)
    offs := runePositions([]byte(testChinese))
    for i := 0; i < b.N; i++ {
        var err error
        reader := rb.BackStreamCount(offs[len(offs)-1], int64(len(offs)))
        for err != io.EOF {
            _, _, err = reader.ReadRune()
        }
    }
    b.SetBytes(int64(len(testChinese)))
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
    offs := runePositions(rb.Value())
    if rb == nil {
        return fmt.Errorf("rope-buffer cannot be nil")
    } else if len(ref) != len(offs) {
        return fmt.Errorf("expected %v runes, got %v", len(ref), len(offs))
    } else if ref != nil {
        // verify length
        if runes, err := rb.Runes(); err != nil {
            return err
        } else if x, y := len(ref), len(runes); x != y {
            return fmt.Errorf("expecting length %d, got %d", x, y)
        }
        // verify value
        if x, y := string(rb.Value()), string(ref); x != y {
            return fmt.Errorf("expecting value %q, got %q", x, y)
        }
        // verify index
        for i := range offs {
            off := offs[i]
            if y, size, err := rb.RuneAt(int64(off)); err != nil {
                return err
            } else if size == 0 {
                return fmt.Errorf("expecting rune at %d for %q", off, string(ref))
            } else if x := ref[i]; x != y {
                return fmt.Errorf("expecting %v, got %v at %d", x, y, off)
            }
        }
        // out of bound index
        if _, _, err := rb.RuneAt(-1); err != ErrorIndexOutofbound {
            return fmt.Errorf("expecting ErrorIndexOutofbound at -1")
        }
        // out of bound index
        _, _, err := rb.RuneAt(rb.Len)
        if err != ErrorIndexOutofbound {
            return fmt.Errorf("expecting ErrorIndexOutofbound at %d", rb.Len)
        }
        // verify substr
        for i, dot := range offs {
            for n := 0; n < len(offs)-i; n++ {
                runes, _, err := rb.RuneSlice(int64(dot), int64(n))
                if err != nil {
                    return err
                }
                if string(runes) != string(ref[i:i+n]) {
                    msg := "expecting %q, got %q"
                    return fmt.Errorf(msg, string(runes), string(ref[dot:n]))
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
        if err := validateRead(rb.(*RopeBuffer), nil); err != nil {
            fmt.Errorf("validateRead() %q at %d: %v", string(x), dot, err)
        }
    }
    _, _, err := rbRef.Split(lenx)
    if err == nil {
        fmt.Errorf("expecting error splitting %q at %d", string(x), lenx)
    }
    return nil
}

func validateRuneSlice(rb *RopeBuffer) error {
    lb := NewLinearBuffer(rb.Value())
    offs := runePositions(rb.Value())
    argList := make([][2]int64, 0)
    for i, dot := range offs {
        for n := 0; n < (len(offs) - i); n++ {
            argList = append(argList, [2]int64{int64(dot), int64(n)})
        }
    }
    for _, arg := range argList {
        dot, n := arg[0], arg[1]
        rbRunes, _, err := rb.RuneSlice(dot, n)
        if err != nil {
            return fmt.Errorf("rb at %d size %d", dot, n)
        }
        lbRunes, _, err := lb.RuneSlice(dot, n)
        if err != nil {
            return fmt.Errorf("lb at %d size %d", dot, n)
        }
        if string(rbRunes) != string(lbRunes) {
            return fmt.Errorf("at %d size %d", dot, n)
        }
    }
    return nil
}
