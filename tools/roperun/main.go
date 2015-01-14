package main

import "io/ioutil"
import "encoding/json"
import "fmt"
import "log"
import "time"
import "github.com/prataprc/monster"
import "github.com/prataprc/goparsec"
import monstc "github.com/prataprc/monster/common"
import "github.com/prataprc/v"

var options {
    prodfile string
    bagdir   string
    seed     int
}

func argParse() {
    prodfile := "../../testdata/rope_test.prod"
    bagdir := "../../testdata"
    seed := uint64(time.Now().UnixNano())

    flag.StringVar(&options.prodfile, "prodfile", prodfile,
        "monster production file to generate commands")
    flag.StringVar(&options.bagdir, "bagdir", bagdir,
        "monster bag dir containing sample texts")
    flag.IntVar(&options.seed, "seed", seed,
        "seed to monster")
}

func main() {
    // read production-file
    text, err := ioutil.ReadFile(options.prodfile)
    if err != nil {
        log.Fatal(err)
    }
    // compile
    root, _ := monster.Y(parsec.NewScanner(text))
    scope := root.(monstc.Scope)
    nterms := scope["_nonterminals"].(monstc.NTForms)
    scope = monster.BuildContext(scope, options.seed, options.bagdir)
    scope["_prodfile"] = options.prodfile

    lb := rope.NewLinearBuffer([]rune(""))
    rb := rope.NewRopebuffer([]rune(""), rope.RopeBufferCapacity)
    // evaluate
    scope = scope.ApplyGlobalForms()
    val := monster.EvalForms("root", scope, nterms["s"])
    lb, rb, err = testCommands(val.(string), lb, rb)
    if err != nil {
        log.Fatalf("seed: %v error: %v\n", options.seed, err)
    }
}

func testCommands(s string,
    lb LinearBuffer, rb *RopeBuffer) (LinearBuffer, *RopeBuffer, error) {

    var cmds []interface{}

    err := json.Unmarshal([]byte(s), &cmds)
    if err != nil {
        return lb, rb, err
    }
    for _, cmd := range cmds {
        lb, rb, err = testCommand(cmd.([]interface{}), lb, rb)
        if err != nil {
            return lb, rb, err
        }
    }
    return lb, rb, nil
}

func testCommand(cmd []interface{},
    lb LinearBuffer, rb *RopeBuffer) (LinearBuffer, *RopeBuffer, error) {

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

func testInsert(dot int64, text []rune,
    lb LinearBuffer, rb *RopeBuffer) (LinearBuffer, *RopeBuffer, error) {

    lb = lb.Insert(dot, text, true)
    rb = rb.Insert(dot, text, true)
    if x, y := string(lb.Value()), string(rb.Value()); x != y {
        return lb, rb, fmt.Errorf("mismatch in input %s : %s", x, y)
    }
    return lb, rb, nil
}

func testDelete(dot, size int64,
    lb LinearBuffer, rb *RopeBuffer) (LinearBuffer, *RopeBuffer, error) {

    //lb = lb.Delete(dot, size)
    //rb = rb.Delete(dot, size)
    //if x, y := string(lb.Value()), string(rb.Value()); x != y {
    //    return lb, rb, fmt.Errorf("mismatch in input %s : %s", x, y)
    //}
    return lb, rb, nil
}

func testIndex(dot int64,
    lb LinearBuffer, rb *RopeBuffer) (LinearBuffer, *RopeBuffer, error) {

    return lb, rb, nil
}

func testLength(
    lb LinearBuffer, rb *RopeBuffer) (LinearBuffer, *RopeBuffer, error) {

    return lb, rb, nil
}

func testValue(
    lb LinearBuffer, rb *RopeBuffer) (LinearBuffer, *RopeBuffer, error) {

    return lb, rb, nil
}

func testSubstr(dot, size int64,
    lb LinearBuffer, rb *RopeBuffer) (LinearBuffer, *RopeBuffer, error) {

    return lb, rb, nil
}
