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
}

//Config represents the command configuration
type Config struct {
	Filename    string
	CmdName     string
	CmdArgs     []string
	QuietMode   bool
	Interactive bool
	PrintHelp   bool
}

const defaultFileName = "termcording"

var cli *flag.FlagSet

//Cmd sets the termcording's cmd
func Cmd(c *exec.Cmd) func(*Termcording) error {
	return func(tc *Termcording) error {
		tc.cmd = c
		return nil
	}
}

//Output sets the termcording's output
func Output(w io.Writer) func(*Termcording) error {
	return func(tc *Termcording) error {
		tc.out = w
		return nil
	}
}

//NewTermcording create a new Termcording instance
func NewTermcording(c *Config, options ...func(*Termcording) error) (*Termcording, error) {
	tc := &Termcording{
		Config: c,
	}
	for _, option := range options {
		err := option(tc)
		if err != nil {
			return &Termcording{}, err
		}
	}
	return tc, nil
}

//TermcordingFromFlags parses command line arguments (and flags) and returns a Config with the values
//of said arguments and flags
func TermcordingFromFlags(options ...func(*Termcording) error) (*Termcording, error) {
	var fName, cmdName string
	var cmdArgs []string

	var h, q bool

	cli = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	cli.BoolVar(&h, "h", false, "Prints this message")
	cli.BoolVar(&q, "q", false, "quite mode - supresses the recording start and end prompts")

	cli.Parse(os.Args[1:])

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
		Filename:    fName,
		CmdName:     cmdName,
		CmdArgs:     cmdArgs,
		QuietMode:   q,
		Interactive: interactive,
		PrintHelp:   h,
	}, options...)
}

//Start creates the script file, creates a new pty and runs the command in that pty
func (tc *Termcording) Start() error {
	if tc.Config.PrintHelp {
		if tc.out == nil {
			tc.out = os.Stdout
		}
		tc.printHelp()
		return nil
	}
	if tc.out == nil {
		f, err := os.OpenFile(tc.Config.Filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			return err
		}
		defer f.Close()
		tc.out = io.MultiWriter(os.Stdout, f)
	}
	if tc.cmd == nil {
		tc.cmd = exec.Command(tc.Config.CmdName, tc.Config.CmdArgs...)
	}
	if !tc.Config.QuietMode {
		fmt.Println("Starting recording session. CTRL-D to end.")
		defer fmt.Printf("\nRecording session ended. Session saved to %s\n", tc.Config.Filename)
	}
	ptmx, restoreMode, err := ptmxFromCmd(tc.cmd, tc.Config.Interactive) //TODO convert to method.
	if err != nil {
		return err
	}

	//inputMWriter := io.MultiWriter(ptmx, f) //TODO: Add option to record stdin as well (as with script -k)
	defer func() {
		restoreMode()
		ptmx.Close()
	}()

	go func() {
		io.Copy(ptmx, os.Stdin)
	}()

	io.Copy(tc.out, ptmx)

	return tc.cmd.Wait()
}

func ptmxFromCmd(c *exec.Cmd, interactive bool) (*os.File, func(), error) {
	modeRestoreFn := func() {}

	ptmx, err := pty.Start(c)
	if err != nil {
		return nil, nil, err
	}
	if interactive {
		// Handle pty size.
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGWINCH)
		go func() {
			for range ch {
				if err := pty.InheritSize(os.Stdin, ptmx); err != nil {
					log.Printf("error resizing pty: %s", err)
				}
			}
		}()
		ch <- syscall.SIGWINCH // Initial resize.

		// Set stdin in raw mode.
		oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			log.Fatalf("failed to set input to raw mode: %s", err)
		}
		modeRestoreFn = func() { term.Restore(int(os.Stdin.Fd()), oldState) }
	}
	return ptmx, modeRestoreFn, err
}

func (tc *Termcording) printHelp() {
	fmt.Fprintf(tc.out, "termcord is a terminal session recorder written in Go.\n\n")
	fmt.Fprintf(tc.out, "Usage: %s [options] [filename [command...]]\n", os.Args[0])
	fmt.Fprintf(tc.out, "Options:\n")
	fmt.Fprintf(tc.out, "-h	Prints this message\n")
	fmt.Fprintf(tc.out, "-q	Quite mode - supresses the recording start and end prompts\n")
	//TODO print default output to tc.writer
	//cli.PrintDefaults()
	fmt.Fprintf(tc.out, "\n")
}
