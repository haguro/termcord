package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/haguro/termcord/pkg/tc"
)

const defaultFileName = "termcording"
const envVar = "TERMCORDING"

type request struct {
	flagSet     *flag.FlagSet
	command     string
	args        []string
	help        bool
	quiet       bool
	append      bool
	interactive bool
	logInput    bool
	filename    string
}

func Run(args []string, in io.Reader, out, errOut io.Writer) int {
	r, err := parseFlags(args, errOut)
	if err != nil && true {
		fmt.Fprintf(errOut, "failed to parse command line arguments: %s", err)
		return -1
	}

	if r.help {
		r.flagSet.Usage()
		return 0
	}

	options := []tc.OptionFunc{}

	c := exec.Command(r.command, r.args...)

	if in == os.Stdin && r.interactive {
		options = append(options, tc.RawMode)
		options = append(options, tc.InheritSizeFrom(in.(*os.File)))
	}

	mode := os.O_TRUNC
	if r.append {
		mode = os.O_APPEND
	}
	f, err := os.OpenFile(r.filename, os.O_WRONLY|os.O_CREATE|mode, 0700)
	if err != nil {
		fmt.Fprintf(errOut, "failed to open file: %s", err)
		return -1
	}
	fmt.Fprintf(f, "Recording started on %s\n", time.Now().Format(time.RFC1123))
	defer func() {
		fmt.Fprintf(f, "Recording ended on %s\n", time.Now().Format(time.RFC1123))
		f.Close()
	}()
	options = append(options, tc.WithOutputWriters(out, f))

	if r.logInput {
		options = append(options, tc.WithInputWriters(f))
	}

	if !r.quiet {
		fmt.Fprintln(out, "Starting recording session. CTRL-D to end.")
		defer fmt.Fprintf(out, "Recording session ended. Session saved to %s\n", r.filename)
	}

	//TODO don't start if the env variable is set?
	os.Setenv(envVar, c.String())

	err = tc.Record(c, options...)
	if err != nil {
		fmt.Fprintf(errOut, "failed to start recording: %s", err)
		return -1
	}

	return 0
}

func parseFlags(args []string, errOut io.Writer) (*request, error) {
	r := request{}

	fs := flag.NewFlagSet(args[0], flag.ExitOnError)
	fs.SetOutput(errOut)
	fs.Usage = printHelpFunc(fs)
	fs.BoolVar(&r.help, "h", false, "Prints this message")
	fs.BoolVar(&r.quiet, "q", false, "Quiet mode - suppresses the recording start and end prompts")
	fs.BoolVar(&r.append, "a", false, "Appends to file instead of overwriting it")
	fs.BoolVar(&r.logInput, "k", false, "Log key strokes to file as well")
	fs.BoolVar(&r.interactive, "i", false, "Run command as interactive. Essential when passing an shell executable as the command argument")
	fs.StringVar(&r.filename, "f", defaultFileName, "Sets recording filename")

	err := fs.Parse(args[1:])
	if err != nil {
		return nil, err
	}
	r.flagSet = fs

	switch fs.NArg() {
	case 0:
		r.command = os.Getenv("SHELL")
		r.interactive = true
	case 1:
		r.command = fs.Arg(0)
	default:
		r.command = fs.Arg(0)
		r.args = fs.Args()[1:]
	}
	return &r, nil
}

func printHelpFunc(fs *flag.FlagSet) func() {
	return func() {
		w := fs.Output()
		fmt.Fprint(w, "termcord is a terminal session recorder written in Go.\n\n")
		fmt.Fprint(w, "Usage: termcord [options] [command [arguments...]]\n")
		fmt.Fprint(w, "Options:\n")
		fs.PrintDefaults()
	}
}
