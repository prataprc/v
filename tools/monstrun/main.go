package main

import "io/ioutil"
import "flag"
import "sync"
import "os/signal"
import "syscall"
import "encoding/json"
import "fmt"
import "log"
import "sort"
import "os"
import "time"
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
	par      int64
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
	flag.Int64Var(&options.par, "par", 1,
		"number of routines to execute a batch of commands")

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

// global data structure on which commands are exercised
var inrw sync.RWMutex
var dsmu sync.Mutex
var ds struct {
	lb *buffer.LinearBuffer
	rb *buffer.RopeBuffer
}

func getds(op, w bool) (*buffer.LinearBuffer, *buffer.RopeBuffer) {
	if op && w {
		inrw.Lock()
	} else if op {
		inrw.RLock()
	}
	dsmu.Lock()
	defer dsmu.Unlock()
	lb, rb := ds.lb, ds.rb
	return lb, rb
}

func setds(lb *buffer.LinearBuffer, rb *buffer.RopeBuffer, op, w bool) {
	dsmu.Lock()
	defer dsmu.Unlock()
	ds.lb, ds.rb = lb, rb
	if op && w {
		inrw.Unlock()
	} else if op {
		inrw.RUnlock()
	}
}

var persistValues = make(map[*buffer.RopeBuffer]string)
var wg = new(sync.WaitGroup)

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
	ds.lb, ds.rb = lb, rb

	// spawn `par` number of routines
	routines := make([]chan []string, 0, options.par)
	for i := int64(0); i < options.par; i++ {
		ch := make(chan []string, 1000)
		log.Printf("spawning routine %d\n", i)
		go doCommands(int(i), ch)
		routines = append(routines, ch)
	}

	// gather persisted data-structure
	killGatherer := make(chan bool)
	go func() {
	loop:
		for {
			switch {
			case <-killGatherer:
				break loop
			default:
			}
			//time.Sleep(10 * time.Millisecond)
			lb, rb = getds(true, false)
			persistValues[rb] = string(rb.Value())
			setds(lb, rb, true, false)
		}
	}()

	// generate a set of many commands in batch.
	for n := int64(0); n < options.count; n += 10 {
		commandsList := generateCommands(text, 10)
		for _, routine := range routines {
			routine <- commandsList
			wg.Add(1)
		}
	}

	wg.Wait()
	close(killGatherer)
	time.Sleep(100 * time.Millisecond)

	printStats()
}

func doCommands(num int, ch chan []string) (i int, j int) {
	defer func() {
		if r := recover(); r != nil {
			printStats()
			log.Println(r)
			log.Println(string(debug.Stack()))
			os.Exit(1)
		}
	}()

	var s string
	var cmd interface{}
	for {
		commandsList := <-ch
		for i, s = range commandsList {
			var cmds []interface{}
			err := json.Unmarshal([]byte(s), &cmds)
			if err != nil {
				incrStat(err.Error())
				continue
			}
			var lb *buffer.LinearBuffer
			var rb *buffer.RopeBuffer
			for j, cmd = range cmds {
				statkey := cmd.([]interface{})[0].(string)
				if statkey == "insertin" || statkey == "deletein" {
					lb, rb = getds(true, true)
				} else {
					lb, rb = getds(true, false)
				}

				//log.Printf("start-routine%d {%d,%d} %v %p %p\n", num, i, j, cmd, lb, rb)
				lb, rb = testCommand(cmd.([]interface{}), lb, rb)
				//log.Printf("end-routine%d {%d,%d} %v %p %p\n", num, i, j, cmd, lb, rb)

				if statkey == "insertin" || statkey == "deletein" {
					setds(lb, rb, true, true)
				} else {
					setds(lb, rb, true, false)
				}
				incrStat(statkey)
			}
		}
		wg.Done()
	}
	return
}

func testCommand(
	cmd []interface{},
	lb *buffer.LinearBuffer,
	rb *buffer.RopeBuffer) (*buffer.LinearBuffer, *buffer.RopeBuffer) {

	var err error

	switch cmd[0] {
	case "insert":
		lb, rb, err = testInsert(
			int64(cmd[1].(float64)), []rune(cmd[2].(string)), lb, rb)
	case "insertin":
		lb, rb, err = testInsertIn(
			int64(cmd[1].(float64)), []rune(cmd[2].(string)), lb, rb)
	case "delete":
		lb, rb, err = testDelete(
			int64(cmd[1].(float64)), int64(cmd[2].(float64)), lb, rb)
	case "deletein":
		lb, rb, err = testDeleteIn(
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
	if err != nil {
		log.Printf("cmd failed: %v\n   %v\n", cmd, err)
		printStats()
	}
	return lb, rb
}

func testInsert(
	dot int64, text []rune,
	lb *buffer.LinearBuffer,
	rb *buffer.RopeBuffer) (*buffer.LinearBuffer, *buffer.RopeBuffer, error) {

	lb1, err1 := lb.Insert(dot, text)
	rb1, err2 := rb.Insert(dot, text)
	if err1 != nil || err2 != nil {
		if err2 != nil {
			incrStat(err2.Error())
		}
		if err1 != err2 {
			return nil, nil, fmt.Errorf("insert: mismatch err %s %s", err1, err2)
		} else if err1 != nil {
			return lb, rb, nil
		}
	}
	if x, y := string(lb1.Value()), string(rb1.Value()); x != y {
		return lb1, rb1, fmt.Errorf("insert: mismatch in text %s : %s", x, y)
	}
	return lb1, rb1, nil
}

func testInsertIn(
	dot int64, text []rune,
	lb *buffer.LinearBuffer,
	rb *buffer.RopeBuffer) (*buffer.LinearBuffer, *buffer.RopeBuffer, error) {

	lb1, err1 := lb.InsertIn(dot, text)
	rb1, err2 := rb.InsertIn(dot, text)
	if err1 != nil || err2 != nil {
		if err2 != nil {
			incrStat(err2.Error())
		}
		if err1 != err2 {
			x, _ := lb.Length()
			y, _ := rb.Length()
			err := fmt.Errorf("insin: mismatch err %v %v %v %v", err1, err2, x, y)
			return nil, nil, err
		} else if err1 != nil {
			return lb, rb, nil
		}
	}
	if x, y := string(lb1.Value()), string(rb1.Value()); x != y {
		return lb1, rb1, fmt.Errorf("insin: mismatch in text %s : %s", x, y)
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
		if err2 != nil {
			incrStat(err2.Error())
		}
		if err1 != err2 {
			x, _ := lb.Length()
			y, _ := rb.Length()
			err := fmt.Errorf("del: mismatch err %v %v %v %v", err1, err2, x, y)
			return nil, nil, err
		} else if err1 != nil {
			return lb, rb, nil
		}
	}
	if x, y := string(lb1.Value()), string(rb1.Value()); x != y {
		return lb1, rb1, fmt.Errorf("mismatch in input %s : %s", x, y)
	}
	return lb1, rb1, nil
}

func testDeleteIn(
	dot, size int64,
	lb *buffer.LinearBuffer,
	rb *buffer.RopeBuffer) (*buffer.LinearBuffer, *buffer.RopeBuffer, error) {

	lb1, err1 := lb.DeleteIn(dot, size)
	rb1, err2 := rb.DeleteIn(dot, size)
	if err1 != nil || err2 != nil {
		if err2 != nil {
			incrStat(err2.Error())
		}
		if err1 != err2 {
			x, _ := lb.Length()
			y, _ := rb.Length()
			err := fmt.Errorf("delin: mismatch err %v %v %v %v", err1, err2, x, y)
			return nil, nil, err
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
		if err2 != nil {
			incrStat(err2.Error())
		}
		if err1 != err2 {
			return nil, nil, fmt.Errorf("runeat: mismatch err %v, %v", err1, err2)
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
		if err2 != nil {
			incrStat(err2.Error())
		}
		if err1 != err2 {
			err := fmt.Errorf("len: mismatch err %v %v %v %v", err1, err2, l1, l2)
			return nil, nil, err
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
		if err2 != nil {
			incrStat(err2.Error())
		}
		if err1 != err2 {
			return nil, nil, fmt.Errorf("subs: mismatch err %v, %v", err1, err2)
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
	"insertin":                          int64(0),
	"delete":                            int64(0),
	"deletein":                          int64(0),
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

func printStats() {
	log.Printf("seed: %v\n", options.seed)
	// verify persistant values.
	log.Printf("verifying %v persistant values ...\n", len(persistValues))
	for rb, refstr := range persistValues {
		if refstr != string(rb.Value()) {
			log.Println("Mismatch ...\n%s\n%s\n\n", refstr, string(rb.Value()))
		}
	}

	// format and print statistics
	total := int64(0)
	for _, val := range stats {
		total += val
	}
	log.Printf("total: %v\n", total)

	keys := make([]string, 0, len(stats))
	for k := range stats {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	log.Println("Stats:")
	for _, k := range keys {
		log.Printf("    %v: %v\n", k, stats[k])
	}
}

func generateCommands(text []byte, count int) []string {
	commandsList := make([]string, count)
	// compile
	root, _ := monster.Y(parsec.NewScanner(text))
	for i := 0; i < count; i++ {
		scope := root.(monstc.Scope)
		nterms := scope["_nonterminals"].(monstc.NTForms)
		scope = monster.BuildContext(scope, options.seed, options.bagdir)
		scope["_prodfile"] = options.prodfile

		// evaluate
		scope = scope.ApplyGlobalForms()
		val := monster.EvalForms("root", scope, nterms["s"])
		commandsList[i] = val.(string)
	}
	return commandsList
}

func signalCatcher() {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT)
	<-ch
	log.Println("CTRL-C; exiting")
	printStats()
	os.Exit(0)
}
