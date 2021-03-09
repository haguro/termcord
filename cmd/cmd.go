package cmd

import (
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
}

//Run creates the script file, creates a new pty and runs the command in that pty
func Run(c *exec.Cmd, f io.Writer, config Config) error {
	if !config.QuietMode {
		fmt.Println("Starting recording session. Use CTRL-D to end.")
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
func ParseArgs() Config {
	var fName, cmdName string
	var cmdArgs []string

	var h, q bool
	var foo string
	flag.BoolVar(&h, "h", false, "Prints this message")
	flag.BoolVar(&q, "q", false, "TODO")
	flag.StringVar(&foo, "f", "", "TODO")
	//TODO moar flags
	flag.Parse()
	if h {
		printHelpAndQuit()
	}

	interactive := true
	switch flag.NArg() {
	case 0:
		fName = "termcording"
		cmdName = getShell()
	case 1:
		fName = flag.Arg(0)
		cmdName = getShell()
	default:
		fName = flag.Arg(0)
		cmdName = flag.Arg(1)
		cmdArgs = flag.Args()[2:]
		interactive = false
	}

	return Config{
		Filename:    fName,
		CmdName:     cmdName,
		CmdArgs:     cmdArgs,
		QuietMode:   q,
		Interactive: interactive,
	}
}

func ptmxFromCmd(c *exec.Cmd, interactive bool) (*os.File, func(), error) {
	ptmx, err := pty.Start(c)
	if err != nil {
		return nil, nil, err
	}
	modeRestoreFn := func() {}
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

func getShell() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		log.Fatal("shell is not set")
	}
	return shell
}

func printHelpAndQuit() {
	// fmt.Println("TODO: Intro")
	// fmt.Println()
	fmt.Printf("Usage: %s [options]\n", os.Args[0])
	fmt.Println("Options:")
	flag.PrintDefaults()
	os.Exit(1)
}
