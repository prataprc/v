package main

import "time"
import "github.com/prataprc/v/term"

func main() {
    err := termbox.Init()
    if err != nil {
        panic(err)
    }
    defer termbox.Close()

    termbox.SetInputMode(termbox.InputEsc | termbox.InputMouse)
    termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
    time.Sleep(1 * time.Second)
}
