package tc

import (
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

// Termcorder holds all values required to record a command run.
type Termcorder struct {
	cmd           *exec.Cmd
	pty           *os.File
	tty           *os.File
	outputWriters io.Writer
	inputWriters  io.Writer
	outputReaders io.Reader
	inputReaders  io.Reader
	prevTermState *term.State
}

// OptionFunc defines a parameter function.
type OptionFunc func(*Termcorder) error

// RawMode sets standard in in raw mode.
func RawMode(t *Termcorder) error {
	// TODO only needed when input is the Standard In. Does this even belong in this package
	// or should it be handled by the caller instead?
	s, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("failed to set Stdin in raw mode: %s", err)
	}
	t.prevTermState = s
	return nil
}

// InheritSizeFrom sets up and returns a parameter function that initially sets the size
// of pty, then launches the resize handler in a go routine.
func InheritSizeFrom(from *os.File) OptionFunc {
	return func(t *Termcorder) error {
		if t.pty == nil {
			return fmt.Errorf("failed to configure term resizing - pty not set")
		}
		err := pty.InheritSize(from, t.pty)
		if err != nil {
			return fmt.Errorf("failed to set pty size: %s", err)
		}
		go setupResizeHandler(from, t.pty)()
		return nil
	}
}

// WithOutputWriters sets up and returns a parameter function that sets the recorder's
// output writers (where the pty's output is copied to).
func WithOutputWriters(writers ...io.Writer) OptionFunc {
	return func(t *Termcorder) error {
		t.outputWriters = io.MultiWriter(writers...)
		return nil
	}
}

// WithInputWriters sets up and returns a parameter function that sets the recorder's
// input writers (what, in addition to the pty, is the input copied to).
func WithInputWriters(writers ...io.Writer) OptionFunc {
	return func(t *Termcorder) error {
		if t.pty == nil {
			return fmt.Errorf("cannot set input writers - pty not set")
		}
		writers = append(writers, t.pty)
		t.inputWriters = io.MultiWriter(writers...)
		return nil
	}
}

// WithInputReaders sets up and returns a parameter function that sets the recorder's
// input readers (what input is copied to the pty).
func WithInputReaders(readers ...io.Reader) OptionFunc {
	return func(t *Termcorder) error {
		t.inputReaders = io.MultiReader(readers...)
		return nil
	}
}

// NewTermcorder creates a new Termcorder instance, setting up and hooking the pty to
// a given command, and configuring readers/writers.
func NewTermcorder(c *exec.Cmd, options ...OptionFunc) (*Termcorder, error) {
	t := Termcorder{
		cmd: c,
	}
	err := t.setupPty()
	if err != nil {
		return nil, err
	}

	//Defaults
	t.outputWriters = os.Stdout
	t.inputWriters = t.pty
	t.outputReaders = t.pty
	t.inputReaders = os.Stdin

	for _, option := range options {
		err := option(&t)
		if err != nil {
			return nil, err
		}
	}
	return &t, nil
}

// Start starts the command in the setup pty, copying readers to writers for both input and output.
func (t *Termcorder) Start() error {
	err := t.cmd.Start()
	if err != nil {
		t.pty.Close()
		return err
	}
	t.tty.Close()
	defer t.pty.Close()
	defer t.restoreTermState()

	go func() {
		io.Copy(t.inputWriters, t.inputReaders)
	}()

	io.Copy(t.outputWriters, t.outputReaders)
	return nil
}

// Record creates a new Termcording and starts it.
func Record(c *exec.Cmd, options ...OptionFunc) error {
	t, err := NewTermcorder(c, options...)
	if err != nil {
		return err
	}
	return t.Start()
}

func (t *Termcorder) restoreTermState() error {
	if t.prevTermState == nil {
		return nil
	}
	return term.Restore(int(os.Stdin.Fd()), t.prevTermState)
}

func (t *Termcorder) setupPty() error {
	ptm, pts, err := pty.Open()
	if err != nil {
		return err
	}

	if t.cmd.SysProcAttr == nil {
		t.cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	t.cmd.SysProcAttr.Setsid = true
	t.cmd.SysProcAttr.Setctty = true

	if t.cmd.Stdin == nil {
		t.cmd.Stdin = pts
	}
	if t.cmd.Stdout == nil {
		t.cmd.Stdout = pts
	}
	if t.cmd.Stderr == nil {
		t.cmd.Stderr = pts
	}

	t.pty = ptm
	t.tty = pts

	return nil
}

func setupResizeHandler(src, dst *os.File) func() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	return func() {
		for range ch {
			if err := pty.InheritSize(src, dst); err != nil {
				log.Printf("error resizing pty: %s", err)
			}
		}
	}
}
