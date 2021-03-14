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
	Config  *Config
	Writers io.Writer
	// Command *exec.Cmd
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

//NewTermcording create a new Termcording instance
func NewTermcording(cfg *Config, w io.Writer) (*Termcording, func() error, error) {
	closerFn := func() error { return nil }
	if w == nil {
		f, err := os.OpenFile(cfg.Filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			return &Termcording{}, nil, err
		}
		closerFn = func() error { return f.Close() }

		w = io.MultiWriter(os.Stdout)
	}

	return &Termcording{
		Writers: w,
		Config:  cfg,
	}, closerFn, nil
}

//Start creates the script file, creates a new pty and runs the command in that pty
func (tc *Termcording) Start() error {
	if tc.Config.PrintHelp {
		printHelp()
		return nil
	}
	if !tc.Config.QuietMode {
		fmt.Println("Starting recording session. CTRL-D to end.")
		defer fmt.Printf("\nRecording session ended. Session saved to %s\n", tc.Config.Filename)
	}

	cmd := exec.Command(tc.Config.CmdName, tc.Config.CmdArgs...)
	ptmx, restoreMode, err := ptmxFromCmd(cmd, tc.Config.Interactive)
	if err != nil {
		return err
	}
	defer func() {
		restoreMode()
		ptmx.Close()
	}()

	//inputMWriter := io.MultiWriter(ptmx, f) //TODO: Add option to record stdin as well (as with script -k)
	go func() {
		io.Copy(ptmx, os.Stdin)
	}()

	io.Copy(tc.Writers, ptmx)

	return cmd.Wait()
}

//TermcordingFromFlags parses command line arguments (and flags) and returns a Config with the values
//of said arguments and flags
func TermcordingFromFlags() (*Termcording, func() error, error) {
	var fName, cmdName string
	var cmdArgs []string

	var h, q bool

	cli = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	cli.BoolVar(&h, "h", false, "Prints this message")
	cli.BoolVar(&q, "q", false, "TODO")

	cli.Parse(os.Args[1:])

	shell, ok := os.LookupEnv("SHELL")
	if cli.Arg(1) == "" && (!ok || shell == "") {
		return &Termcording{}, func() error { return nil }, errors.New("shell not set")
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
	}, nil)
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

func printHelp() {
	fmt.Println("TODO: Intro")
	fmt.Println()
	fmt.Printf("Usage: %s [options] [filename [command...]]\n", os.Args[0])
	fmt.Println("Options:")
	cli.PrintDefaults()
	fmt.Println()
}
