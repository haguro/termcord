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
	"time"

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
	var cmdName, f string
	var cmdArgs []string
	var h, q, a, k, i bool

	cli = flag.NewFlagSet(args[0], flag.ExitOnError)
	cli.BoolVar(&h, "h", false, "Prints this message")
	cli.BoolVar(&q, "q", false, "Quiet mode - suppresses the recording start and end prompts")
	cli.BoolVar(&a, "a", false, "Appends to file instead of overwriting it")
	cli.BoolVar(&k, "k", false, "Log key strokes to file as well")
	cli.BoolVar(&i, "i", false, "Run command as interactive. Essential when passing an shell executable as the command argument")
	cli.StringVar(&f, "f", defaultFileName, "Sets recording filename")

	cli.Parse(args[1:])

	switch cli.NArg() {
	case 0:
		shell := os.Getenv("SHELL")
		if shell == "" {
			return nil, errors.New("shell empty or not set")
		}
		cmdName, i = shell, true
	case 1:
		cmdName = cli.Arg(0)
	default:
		cmdName = cli.Arg(0)
		cmdArgs = cli.Args()[1:]
	}

	return NewTermcording(&Config{
		Filename:      f,
		CmdName:       cmdName,
		CmdArgs:       cmdArgs,
		QuietMode:     q,
		Append:        a,
		Interactive:   i,
		LogKeystrokes: k,
		PrintHelp:     h,
	}, options...)
}

//Start runs the tc.cmd in a pseudo-terminal and writes all output to tc.out
func (tc *Termcording) Start() error {
	if tc.Config.PrintHelp {
		w := tc.out
		if w == nil {
			w = os.Stdout
		}
		printHelp(w)
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
	fmt.Fprintf(f, "Recording started on %s\n", time.Now().Format(time.RFC1123))
	closer = func() error {
		fmt.Fprintf(f, "Recording ended on %s\n", time.Now().Format(time.RFC1123))
		return f.Close()
	}

	if tc.Config.LogKeystrokes {
		tc.in = io.MultiWriter(tc.in, f)
	}
	tc.out = io.MultiWriter(os.Stdout, f)

	return closer, nil
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

func printHelp(w io.Writer) {
	fmt.Fprint(w, "termcord is a terminal session recorder written in Go.\n\n")
	fmt.Fprint(w, "Usage: termcord [options] [command [arguments...]]\n")
	fmt.Fprint(w, "Options:\n")
	cli.VisitAll(func(f *flag.Flag) {
		fmt.Fprintf(w, "  -%s	%s\n", f.Name, f.Usage)
	})
}
