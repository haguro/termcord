package termcorder

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/creack/pty"
	"golang.org/x/term"
)

//Termcording represent a command/terminal recording
type Termcording struct {
	Config *Config
	cmd    *exec.Cmd
	out    io.Writer
	in     io.Writer
}

//Config represents the command configuration
type Config struct {
	Filename      string
	CmdName       string
	CmdArgs       []string
	QuietMode     bool
	Append        bool
	Interactive   bool
	LogKeystrokes bool
	PrintHelp     bool
}

const defaultFileName = "termcording"

var cli *flag.FlagSet

//Cmd sets the termcording's cmd.
func Cmd(c *exec.Cmd) func(*Termcording) error {
	return func(tc *Termcording) error {
		tc.cmd = c
		return nil
	}
}

//Output sets the termcording's output.
func Output(w io.Writer) func(*Termcording) error {
	return func(tc *Termcording) error {
		tc.out = w
		return nil
	}
}

//NewTermcording returns a pointer to a new variable of type Termcording given a config
//variable and (functional) options.
func NewTermcording(c *Config, options ...func(*Termcording) error) (*Termcording, error) {
	tc := &Termcording{
		Config: c,
	}
	for _, option := range options {
		err := option(tc)
		if err != nil {
			return nil, err
		}
	}
	return tc, nil
}

//FromFlags parses command line flags (and arguments) and returns a pointer to a
//new variable of type `Termcording`.
func FromFlags(args []string, options ...func(*Termcording) error) (*Termcording, error) {
	var fName, cmdName string
	var cmdArgs []string

	var h, q, a, k bool

	cli = flag.NewFlagSet(args[0], flag.ExitOnError)
	cli.BoolVar(&h, "h", false, "Prints this message")
	cli.BoolVar(&q, "q", false, "Quiet mode - suppresses the recording start and end prompts")
	cli.BoolVar(&a, "a", false, "Appends to file instead of overwriting it")
	cli.BoolVar(&k, "k", false, "Log key strokes to file as well")

	cli.Parse(args[1:])

	shell, ok := os.LookupEnv("SHELL")
	if cli.Arg(1) == "" && (!ok || shell == "") {
		return &Termcording{}, errors.New("shell not set")
	}
	interactive := true
	switch cli.NArg() {
	case 0:
		fName = defaultFileName
		cmdName = shell
	case 1:
		fName = cli.Arg(0)
		cmdName = shell
	default:
		fName = cli.Arg(0)
		cmdName = cli.Arg(1)
		cmdArgs = cli.Args()[2:]
		interactive = false
	}

	return NewTermcording(&Config{
		Filename:      fName,
		CmdName:       cmdName,
		CmdArgs:       cmdArgs,
		QuietMode:     q,
		Append:        a,
		Interactive:   interactive,
		LogKeystrokes: k,
		PrintHelp:     h,
	}, options...)
}

//Start runs the tc.cmd in a pseudo-terminal and writes all output to tc.out
func (tc *Termcording) Start() error {
	if tc.Config.PrintHelp {
		if tc.out == nil {
			tc.out = os.Stdout
		}
		printHelp(tc)
		return nil
	}

	if tc.cmd == nil {
		tc.cmd = exec.Command(tc.Config.CmdName, tc.Config.CmdArgs...)
	}

	if !tc.Config.QuietMode && tc.out == nil {
		fmt.Println("Starting recording session. CTRL-D to end.")
		defer fmt.Printf("\nRecording session ended. Session saved to %s\n", tc.Config.Filename)
	}

	pterm, restoreMode, err := pseudoTermFromCmd(tc.cmd, tc.Config.Interactive)
	if err != nil {
		return fmt.Errorf("failed to create a pty from command: %s", err)
	}
	defer func() {
		restoreMode()
		pterm.Close()
	}()

	tc.in = io.Writer(pterm)
	if tc.out == nil {
		closerFn, err := tc.setupWriters()
		if err != nil {
			return fmt.Errorf("failed to set up recording outputs: %s", err)
		}
		defer closerFn()
	}

	go func() {
		io.Copy(tc.in, os.Stdin)
	}()

	io.Copy(tc.out, pterm)

	return tc.cmd.Wait()
}

func (tc *Termcording) setupWriters() (closer func() error, err error) {
	mode := os.O_TRUNC
	if tc.Config.Append {
		mode = os.O_APPEND
	}
	f, err := os.OpenFile(tc.Config.Filename, os.O_WRONLY|os.O_CREATE|mode, 0700)
	if err != nil {
		return nil, fmt.Errorf("failed to open recording file: %s", err)
	}
	if tc.Config.LogKeystrokes {
		tc.in = io.MultiWriter(tc.in, f)
	}
	tc.out = io.MultiWriter(os.Stdout, f)
	return f.Close, nil
}

func pseudoTermFromCmd(c *exec.Cmd, interactive bool) (pterm *os.File, stdinModeRestore func(), err error) {
	var s *pty.Winsize
	if interactive {
		s, err = pty.GetsizeFull(os.Stdin)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get terminal size: %s", err)
		}
	}
	pterm, err = pty.StartWithSize(c, s)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start the pty process: %s", err)
	}

	stdinModeRestore = func() {}
	if interactive {
		// Handle pty size.
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGWINCH)
		go func() {
			for range ch {
				if err := pty.InheritSize(os.Stdin, pterm); err != nil {
					log.Printf("error resizing pty: %s", err)
				}
			}
		}()

		// Set stdin in raw mode.
		oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			return nil, nil, fmt.Errorf("failed to set Stdin in raw mode: %s", err)
		}
		stdinModeRestore = func() { term.Restore(int(os.Stdin.Fd()), oldState) }
	}
	return pterm, stdinModeRestore, err
}

func printHelp(tc *Termcording) {
	fmt.Fprint(tc.out, "termcord is a terminal session recorder written in Go.\n\n")
	fmt.Fprint(tc.out, "Usage: termcord [options] [filename [command...]]\n")
	fmt.Fprint(tc.out, "Options:\n")
	cli.VisitAll(func(f *flag.Flag) {
		fmt.Fprintf(tc.out, "  -%s	%s\n", f.Name, f.Usage)
	})
}
