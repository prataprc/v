package v

// h means {0,z} select everyting.

type Cmd struct {
    args []interface{}
    c    string
    zrgs []interface{}
}

// Commands
const (
    // insert mode.
    RuboutChar string = "k"
    RuboutWord        = "ctrl-w"
    RuboutLine        = "ctrl-u"
    // normal mode.
    DotForward    = "l" // {args{Int}, "l"}
    DotForwardTok = "w" // {args{Int}, "w"}
    DotLineUp     = "k" // {args{Int}, "k"}
    DotGoto       = "j" // {args{Int}, "j"}
    // ex-command.
    Ex = "exit"
)
