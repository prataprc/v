package main

import "io/ioutil"
import "flag"
import "encoding/json"
import "fmt"
import "log"
import "os"
import "time"

import "github.com/prataprc/monster"
import "github.com/prataprc/goparsec"
import monstc "github.com/prataprc/monster/common"
import ds "github.com/prataprc/v/rope"

var options struct {
	prodfile string
	bagdir   string
	seed     uint64
	count    uint64
}

func argParse() {
	seed := uint64(time.Now().UnixNano())

	flag.StringVar(&options.prodfile, "prodfile", "",
		"monster production file to generate commands")
	flag.StringVar(&options.bagdir, "bagdir", "",
		"monster bag dir containing sample texts")
	flag.Uint64Var(&options.seed, "seed", seed,
		"seed to monster")
	flag.Uint64Var(&options.count, "count", 1,
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

func main() {
	argParse()
	// read production-file
	text, err := ioutil.ReadFile(options.prodfile)
	if err != nil {
		log.Fatal(err)
	}
	stats := make(map[string]int64)
	for options.count > 0 {
		// compile
		root, _ := monster.Y(parsec.NewScanner(text))
		scope := root.(monstc.Scope)
		nterms := scope["_nonterminals"].(monstc.NTForms)
		scope = monster.BuildContext(scope, options.seed, options.bagdir)
		scope["_prodfile"] = options.prodfile

		lb := ds.NewLinearBuffer([]rune(""))
		rb := ds.NewRopebuffer([]rune(""), ds.RopeBufferCapacity)
		// evaluate
		scope = scope.ApplyGlobalForms()
		val := monster.EvalForms("root", scope, nterms["s"])
		lb, rb, stats, err = testCommands(val.(string), lb, rb, stats)
		if err != nil {
			log.Fatalf("seed: %v error: %v\n", options.seed, err)
		}
		options.count--
	}
	fmt.Println(stats)
}

func testCommands(
	s string,
	lb *ds.LinearBuffer,
	rb *ds.RopeBuffer,
	stats map[string]int64) (*ds.LinearBuffer, *ds.RopeBuffer, map[string]int64, error) {

	var cmds []interface{}

	err := json.Unmarshal([]byte(s), &cmds)
	if err != nil {
		return lb, rb, stats, err
	}
	for _, cmd := range cmds {
		statkey := cmd.([]interface{})[0].(string)
		val, ok := stats[statkey]
		if !ok {
			val = 0
		}
		lb, rb, err = testCommand(cmd.([]interface{}), lb, rb)
		if err != nil {
			return lb, rb, stats, err
		}
		stats[statkey] = val + 1
	}
	return lb, rb, stats, nil
}

func testCommand(
	cmd []interface{},
	lb *ds.LinearBuffer,
	rb *ds.RopeBuffer) (*ds.LinearBuffer, *ds.RopeBuffer, error) {

	var err error
	switch cmd[0] {
	case "insert":
		lb, rb, err = testInsert(
			int64(cmd[1].(float64)), []rune(cmd[2].(string)), lb, rb)
	case "delete":
		lb, rb, err = testDelete(
			int64(cmd[1].(float64)), int64(cmd[2].(float64)), lb, rb)
	case "index":
		lb, rb, err = testIndex(int64(cmd[1].(float64)), lb, rb)
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
	lb *ds.LinearBuffer,
	rb *ds.RopeBuffer) (*ds.LinearBuffer, *ds.RopeBuffer, error) {

	lb1, err1 := lb.Insert(dot, text)
	rb1, err2 := rb.Insert(dot, text)
	if err1 != err2 {
		return nil, nil, fmt.Errorf("mismatch in err %s %s", err1, err2)
	} else if err1 != nil {
		return lb, rb, nil
	}
	if x, y := string(lb1.Value()), string(rb1.Value()); x != y {
		return lb1, rb1, fmt.Errorf("mismatch in input %s : %s", x, y)
	}
	return lb1, rb1, nil
}

func testDelete(
	dot, size int64,
	lb *ds.LinearBuffer,
	rb *ds.RopeBuffer) (*ds.LinearBuffer, *ds.RopeBuffer, error) {

	lb1, err1 := lb.Delete(dot, size)
	rb1, err2 := rb.Delete(dot, size)
	if err1 != err2 {
		return nil, nil, fmt.Errorf("mismatch in err %s %s", err1, err2)
	} else if err1 != nil {
		return lb, rb, nil
	}
	if x, y := string(lb1.Value()), string(rb1.Value()); x != y {
		return lb1, rb1, fmt.Errorf("mismatch in input %s : %s", x, y)
	}
	return lb1, rb1, nil
}

func testIndex(
	dot int64,
	lb *ds.LinearBuffer,
	rb *ds.RopeBuffer) (*ds.LinearBuffer, *ds.RopeBuffer, error) {

	ch1, ok1, err1 := lb.Index(dot)
	ch2, ok2, err2 := rb.Index(dot)
	if err1 != err2 {
		return nil, nil, fmt.Errorf("mismatch in err %v, %v", err1, err2)
	} else if ok1 != ok2 {
		return nil, nil, fmt.Errorf("mismatch in ok %v, %v", ok1, ok2)
	} else if ch1 != ch2 {
		return nil, nil, fmt.Errorf("mismatch in ch %v, %v", ch1, ch2)
	}
	return lb, rb, nil
}

func testLength(
	lb *ds.LinearBuffer,
	rb *ds.RopeBuffer) (*ds.LinearBuffer, *ds.RopeBuffer, error) {

	l1, err1 := lb.Length()
	l2, err2 := rb.Length()
	if err1 != err2 {
		return nil, nil, fmt.Errorf("mismatch in err %v, %v", err1, err2)
	} else if l1 != l2 {
		return nil, nil, fmt.Errorf("mismatch in len %v, %v", l1, l2)
	}
	return lb, rb, nil
}

func testValue(
	lb *ds.LinearBuffer,
	rb *ds.RopeBuffer) (*ds.LinearBuffer, *ds.RopeBuffer, error) {

	if x, y := string(lb.Value()), string(rb.Value()); x != y {
		return nil, nil, fmt.Errorf("mismatch in value %v, %v", x, y)
	}
	return lb, rb, nil
}

func testSubstr(
	dot,
	size int64,
	lb *ds.LinearBuffer,
	rb *ds.RopeBuffer) (*ds.LinearBuffer, *ds.RopeBuffer, error) {

	s1, err1 := lb.Substr(dot, size)
	s2, err2 := rb.Substr(dot, size)
	if err1 != err2 {
		return nil, nil, fmt.Errorf("mismatch in err %v, %v", err1, err2)
	} else if s1 != s2 {
		return nil, nil, fmt.Errorf("mismatch in substr %v, %v", s1, s2)
	}
	return lb, rb, nil
}
