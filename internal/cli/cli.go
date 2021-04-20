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

// DefaultFileName sets the default name of the recording file.
const DefaultFileName = "termcording"

// EnvVar sets the key of the environment variable that is set while the command is executed.
const EnvVar = "TERMCORDING"

type request struct {
	flagSet       *flag.FlagSet
	command       string
	args          []string
	input         io.Reader
	output        io.Writer
	errorOutput   io.Writer
	recorderSetup RecorderSetupFunc
	help          bool
	quiet         bool
	append        bool
	interactive   bool
	logInput      bool
	filename      string
}

// RecorderSetupFunc defines the functions that sets up the recording destination.
// A `RecorderSetupFunc` function is passed to and subsequently executed by `Run` set up the recorder.
// This could be anything from a file on the file system, a network socket, or a simple bytes buffer.
type RecorderSetupFunc func(string, bool) (io.ReadWriteCloser, error)

// FileRecorderSetup sets up the recording file as the recording destination.
func FileRecorderSetup(filename string, append bool) (file io.ReadWriteCloser, err error) {
	mode := os.O_TRUNC
	if append {
		mode = os.O_APPEND
	}
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|mode, 0700)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// WithInput returns a function that sets the input of the input of the request.
func WithInput(input io.Reader) func(r *request) {
	return func(r *request) {
		r.input = input
	}
}

// WithOutput returns a function that sets the output of the input of the request.
func WithOutput(output io.Writer) func(r *request) {
	return func(r *request) {
		r.output = output
	}
}

// WithErrorOutput returns a function that sets the error output of the input of the request.
func WithErrorOutput(errorOut io.Writer) func(r *request) {
	return func(r *request) {
		r.errorOutput = errorOut
	}
}

// WithRecorderSetupFunc returns a function that sets the recorder setup function of the request.
func WithRecorderSetupFunc(f RecorderSetupFunc) func(r *request) {
	return func(r *request) {
		r.recorderSetup = f
	}
}

// Run parses the command arguments, sets up the recording destination and initiates the terminal recorder.
func Run(args []string, options ...func(r *request)) int {
	r := request{
		input:         os.Stdin,
		output:        os.Stdout,
		errorOutput:   os.Stderr,
		recorderSetup: FileRecorderSetup,
	}
	for _, option := range options {
		option(&r)
	}

	err := r.parseFlags(args)
	if err != nil {
		fmt.Fprintf(r.errorOutput, "failed to parse command line arguments: %s", err)
		return -1
	}

	if r.help {
		r.flagSet.Usage()
		return 0
	}

	//TODO don't start if envVar is set?

	tcOptions := []tc.OptionFunc{}

	c := exec.Command(r.command, r.args...)

	if r.input == os.Stdin && r.interactive {
		tcOptions = append(tcOptions, tc.RawMode)
		tcOptions = append(tcOptions, tc.InheritSizeFrom(os.Stdin))
	}

	f, err := r.recorderSetup(r.filename, r.append)
	if err != nil {
		fmt.Fprintf(r.errorOutput, "failed to set up recorder: %s", err)
		return -1
	}
	fmt.Fprintf(f, "Recording started on %s\n", time.Now().Format(time.RFC1123))
	defer func() {
		fmt.Fprintf(f, "Recording ended on %s\n", time.Now().Format(time.RFC1123))
		f.Close()
	}()
	tcOptions = append(tcOptions, tc.WithOutputWriters(r.output, f))

	if r.logInput {
		tcOptions = append(tcOptions, tc.WithInputWriters(f))
	}

	if !r.quiet {
		fmt.Fprintln(r.output, "Starting recording session. CTRL-D to end.")
		defer fmt.Fprintf(r.output, "Recording session ended. Session saved to %s\n", r.filename)
	}

	os.Setenv(EnvVar, c.String())

	err = tc.Record(c, tcOptions...)
	if err != nil {
		fmt.Fprintf(r.errorOutput, "failed to start recording: %s", err)
		return -1
	}

	return 0
}

func (r *request) parseFlags(args []string) error {
	fs := flag.NewFlagSet(args[0], flag.ExitOnError)
	fs.SetOutput(r.errorOutput)
	fs.Usage = printHelpFunc(fs)
	fs.BoolVar(&r.help, "h", false, "Prints this message")
	fs.BoolVar(&r.quiet, "q", false, "Quiet mode - suppresses the recording start and end prompts")
	fs.BoolVar(&r.append, "a", false, "Appends to file instead of overwriting it")
	fs.BoolVar(&r.logInput, "k", false, "Log key strokes to file as well")
	fs.BoolVar(&r.interactive, "i", false, "Run command as interactive. Essential when passing a shell executable as the command argument")
	fs.StringVar(&r.filename, "f", DefaultFileName, "Sets recording filename")

	err := fs.Parse(args[1:])
	if err != nil {
		return err
	}
	r.flagSet = fs

	if r.help {
		return nil
	}

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
	return nil
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
