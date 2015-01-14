package rope

import "testing"
import "io/ioutil"
import "fmt"

var _ = fmt.Sprintf("dummy")

func TestRopeSample(t *testing.T) {
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
    fmt.Println(stats)
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

