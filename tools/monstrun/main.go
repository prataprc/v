package main

import "io/ioutil"
import "flag"
import "encoding/json"
import "fmt"
import "log"
import "os"
import "time"
import "sync"
import "reflect"
import "runtime/debug"

import "github.com/prataprc/monster"
import "github.com/prataprc/goparsec"
import monstc "github.com/prataprc/monster/common"
import "github.com/prataprc/v/buffer"

var options struct {
	prodfile string
	bagdir   string
	seed     uint64
	count    int64
}

func argParse() {
	seed := uint64(time.Now().UnixNano())

	flag.StringVar(&options.prodfile, "prodfile", "",
		"monster production file to generate commands")
	flag.StringVar(&options.bagdir, "bagdir", "",
		"monster bag dir containing sample texts")
	flag.Uint64Var(&options.seed, "seed", seed,
		"seed to monster")
	flag.Int64Var(&options.count, "count", 1,
		"loop count to run monster")

	flag.Parse()

	if options.prodfile == "" {
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage : <prog> [OPTIONS] \n")
	flag.PrintDefaults()
}

var dsmu sync.Mutex
var ds struct {
	lb *buffer.LinearBuffer
	rb *buffer.RopeBuffer
}

func getds() (*buffer.LinearBuffer, *buffer.RopeBuffer) {
	dsmu.Lock()
	defer dsmu.Unlock()
	lb, rb := ds.lb, ds.rb
	return lb, rb
}

func setds(lb *buffer.LinearBuffer, rb *buffer.RopeBuffer) {
	dsmu.Lock()
	defer dsmu.Unlock()
	ds.lb, ds.rb = lb, rb
}

func main() {
	argParse()
	// read production-file
	text, err := ioutil.ReadFile(options.prodfile)
	if err != nil {
		log.Fatal(err)
	}
	lb := buffer.NewLinearBuffer([]byte(""))
	rb, err := buffer.NewRopebuffer([]byte(""), buffer.RopeBufferCapacity)
	if err != nil {
		log.Fatal(err)
	}
	setds(lb, rb)
	ch := make(chan bool, options.count)
	n := int64(0)
	for n < options.count {
		// compile
		root, _ := monster.Y(parsec.NewScanner(text))
		scope := root.(monstc.Scope)
		nterms := scope["_nonterminals"].(monstc.NTForms)
		scope = monster.BuildContext(scope, options.seed, options.bagdir)
		scope["_prodfile"] = options.prodfile

		// evaluate
		scope = scope.ApplyGlobalForms()
		val := monster.EvalForms("root", scope, nterms["s"])
		go testCommands(val.(string), ch)
		if err != nil {
			log.Fatalf("seed: %v error: %v\n", options.seed, err)
		}
		n++
	}
	// wait until all tests are executed
	n = int64(0)
	persistValues := make(map[*buffer.RopeBuffer]string)
	for n < options.count {
		switch {
		case <-ch:
			n++
		default:
		}
		_, rb = getds()
		persistValues[rb] = string(rb.Value())
	}
	// verify persistant values.
	for rb, refstr := range persistValues {
		if refstr != string(rb.Value()) {
			fmt.Printf("seed: %v\n", options.seed)
			fmt.Println("Mismatch ...\n%s\n%s\n\n", refstr, string(rb.Value()))
		}
	}
	// format and print statistics
	total := int64(0)
	for _, val := range stats {
		total += val
	}
	fmt.Printf("total: %v\n%v\n", total, stats)
	fmt.Printf("verified: %v persistant values\n", len(persistValues))
}

func testCommands(s string, ch chan bool) {
	var cmds []interface{}

	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r, "seed:", options.seed)
			fmt.Println(string(debug.Stack()))
			os.Exit(1)
		}
	}()

	err := json.Unmarshal([]byte(s), &cmds)
	if err != nil {
		incrStat(err.Error())
		return
	}
	for _, cmd := range cmds {
		lb, rb := getds()
		statkey := cmd.([]interface{})[0].(string)
		lb, rb, err = testCommand(cmd.([]interface{}), lb, rb)
		if err != nil {
			fmt.Println(err)
			return
		}
		incrStat(statkey)
		setds(lb, rb)
	}
	ch <- true
	return
}

func testCommand(
	cmd []interface{},
	lb *buffer.LinearBuffer,
	rb *buffer.RopeBuffer) (*buffer.LinearBuffer, *buffer.RopeBuffer, error) {

	var err error
	switch cmd[0] {
	case "insert":
		lb, rb, err = testInsert(
			int64(cmd[1].(float64)), []rune(cmd[2].(string)), lb, rb)
	case "delete":
		lb, rb, err = testDelete(
			int64(cmd[1].(float64)), int64(cmd[2].(float64)), lb, rb)
	case "index":
		lb, rb, err = testRuneAt(int64(cmd[1].(float64)), lb, rb)
	case "length":
		lb, rb, err = testLength(lb, rb)
	case "value":
		lb, rb, err = testValue(lb, rb)
	case "substr":
		lb, rb, err = testSubstr(
			int64(cmd[1].(float64)), int64(cmd[2].(float64)), lb, rb)
	}
	return lb, rb, err
}

func testInsert(
	dot int64, text []rune,
	lb *buffer.LinearBuffer,
	rb *buffer.RopeBuffer) (*buffer.LinearBuffer, *buffer.RopeBuffer, error) {

	lb1, err1 := lb.Insert(dot, text)
	rb1, err2 := rb.Insert(dot, text)
	if err1 != nil || err2 != nil {
		incrStat(err1.Error())
		if err1 != err2 {
			return nil, nil, fmt.Errorf("mismatch in err %s %s", err1, err2)
		} else if err1 != nil {
			return lb, rb, nil
		}
	}
	if x, y := string(lb1.Value()), string(rb1.Value()); x != y {
		return lb1, rb1, fmt.Errorf("mismatch in text %s : %s", x, y)
	}
	return lb1, rb1, nil
}

func testDelete(
	dot, size int64,
	lb *buffer.LinearBuffer,
	rb *buffer.RopeBuffer) (*buffer.LinearBuffer, *buffer.RopeBuffer, error) {

	lb1, err1 := lb.Delete(dot, size)
	rb1, err2 := rb.Delete(dot, size)
	if err1 != nil || err2 != nil {
		incrStat(err1.Error())
		if err1 != err2 {
			return nil, nil, fmt.Errorf("mismatch in err %s %s", err1, err2)
		} else if err1 != nil {
			return lb, rb, nil
		}
	}
	if x, y := string(lb1.Value()), string(rb1.Value()); x != y {
		return lb1, rb1, fmt.Errorf("mismatch in input %s : %s", x, y)
	}
	return lb1, rb1, nil
}

func testRuneAt(
	dot int64,
	lb *buffer.LinearBuffer,
	rb *buffer.RopeBuffer) (*buffer.LinearBuffer, *buffer.RopeBuffer, error) {

	ch1, size1, err1 := lb.RuneAt(dot)
	ch2, size2, err2 := rb.RuneAt(dot)
	if err1 != nil || err2 != nil {
		incrStat(err1.Error())
		if err1 != err2 {
			return nil, nil, fmt.Errorf("mismatch in err %v, %v", err1, err2)
		} else if size1 != size2 {
			return nil, nil, fmt.Errorf("mismatch in size %v, %v", size1, size2)
		} else if ch1 != ch2 {
			return nil, nil, fmt.Errorf("mismatch in ch %v, %v", ch1, ch2)
		}
	}
	return lb, rb, nil
}

func testLength(
	lb *buffer.LinearBuffer,
	rb *buffer.RopeBuffer) (*buffer.LinearBuffer, *buffer.RopeBuffer, error) {

	l1, err1 := lb.Length()
	l2, err2 := rb.Length()
	if err1 != nil || err2 != nil {
		incrStat(err1.Error())
		if err1 != err2 {
			return nil, nil, fmt.Errorf("mismatch in err %v, %v", err1, err2)
		} else if l1 != l2 {
			return nil, nil, fmt.Errorf("mismatch in len %v, %v", l1, l2)
		}
	}
	return lb, rb, nil
}

func testValue(
	lb *buffer.LinearBuffer,
	rb *buffer.RopeBuffer) (*buffer.LinearBuffer, *buffer.RopeBuffer, error) {

	if x, y := string(lb.Value()), string(rb.Value()); x != y {
		return nil, nil, fmt.Errorf("mismatch in value %v, %v", x, y)
	}
	return lb, rb, nil
}

func testSubstr(
	dot,
	size int64,
	lb *buffer.LinearBuffer,
	rb *buffer.RopeBuffer) (*buffer.LinearBuffer, *buffer.RopeBuffer, error) {

	rs1, size1, err1 := lb.RuneSlice(dot, size)
	rs2, size2, err2 := rb.RuneSlice(dot, size)
	if err1 != nil || err2 != nil {
		incrStat(err1.Error())
		if err1 != err2 {
			return nil, nil, fmt.Errorf("mismatch in err %v, %v", err1, err2)
		} else if size1 != size2 {
			return nil, nil, fmt.Errorf("mismatch in size %v, %v", size1, size2)
		} else if !reflect.DeepEqual(rs1, rs2) {
			s1, s2 := string(rs1), string(rs2)
			return nil, nil, fmt.Errorf("mismatch in substr %v, %v", s1, s2)
		}
	}
	return lb, rb, nil
}

// update statistics
var statsmu sync.Mutex
var stats = map[string]int64{
	"insert":                            int64(0),
	"delete":                            int64(0),
	"index":                             int64(0),
	"length":                            int64(0),
	"value":                             int64(0),
	"substr":                            int64(0),
	buffer.ErrorBufferNil.Error():       int64(0),
	buffer.ErrorIndexOutofbound.Error(): int64(0),
}

func incrStat(key string) {
	statsmu.Lock()
	defer statsmu.Unlock()
	stats[key] = stats[key] + 1
}
