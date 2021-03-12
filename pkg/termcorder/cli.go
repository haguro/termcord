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

//Run creates the script file, creates a new pty and runs the command in that pty
func Run(c *exec.Cmd, f io.Writer, config *Config) error {
	if config.PrintHelp {
		printHelp()
		return nil
	}
	if !config.QuietMode {
		fmt.Println("Starting recording session. CTRL-D to end.")
		defer fmt.Printf("\nRecording session ended. Session saved to %s\n", config.Filename)
	}

	ptmx, restoreMode, err := ptmxFromCmd(c, config.Interactive)
	if err != nil {
		return err
	}
	defer func() {
		restoreMode()
		ptmx.Close()
	}()

	//inputMWriter := io.MultiWriter(ptmx, f) //TODO: Add option to record stdin as well (as with script -k)
	outputMWriter := io.MultiWriter(os.Stdout, f)

	go func() {
		io.Copy(ptmx, os.Stdin)
	}()

	io.Copy(outputMWriter, ptmx)

	return c.Wait()
}

//ParseArgs parses command line arguments (and flags) and returns a Config with the values
//of said arguments and flags
func ParseArgs() (*Config, error) {
	var fName, cmdName string
	var cmdArgs []string

	var h, q bool

	cli = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	cli.BoolVar(&h, "h", false, "Prints this message")
	cli.BoolVar(&q, "q", false, "TODO")

	cli.Parse(os.Args[1:])

	shell, ok := os.LookupEnv("SHELL")
	if cli.Arg(1) == "" && (!ok || shell == "") {
		return &Config{}, errors.New("shell not set")
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

	return &Config{
		Filename:    fName,
		CmdName:     cmdName,
		CmdArgs:     cmdArgs,
		QuietMode:   q,
		Interactive: interactive,
		PrintHelp:   h,
	}, nil
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
