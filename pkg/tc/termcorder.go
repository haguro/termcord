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

type Termcorder struct {
	cmd           *exec.Cmd
	pty           *os.File
	outputWriter  io.Writer
	inputWriter   io.Writer
	outputReader  io.Reader
	inputReader   io.Reader
	prevTermState *term.State
}

type OptionFunc func(*Termcorder) error

func RawMode(t *Termcorder) error {
	//TODO only needed when input is the Standard In. Does this even belong in this package or should it be handled by the caller instead?
	s, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("failed to set Stdin in raw mode: %s", err)
	}
	t.prevTermState = s
	return nil
}

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

func WithOutputWriters(writers ...io.Writer) OptionFunc {
	return func(t *Termcorder) error {
		t.outputWriter = io.MultiWriter(writers...)
		return nil
	}
}

func WithInputWriters(writers ...io.Writer) OptionFunc {
	return func(t *Termcorder) error {
		if t.pty == nil {
			return fmt.Errorf("cannot set input writers - pty not set")
		}
		writers = append(writers, t.pty)
		t.inputWriter = io.MultiWriter(writers...)
		return nil
	}
}

func WithInputReaders(readers ...io.Reader) OptionFunc {
	return func(t *Termcorder) error {
		t.inputReader = io.MultiReader(readers...)
		return nil
	}
}

func NewTermcorder(c *exec.Cmd, options ...OptionFunc) (*Termcorder, error) {
	t := Termcorder{
		cmd: c,
	}
	err := t.setupPty()
	if err != nil {
		return nil, err
	}

	//Defaults
	t.outputWriter = os.Stdout
	t.inputWriter = t.pty
	t.outputReader = t.pty
	t.inputReader = os.Stdin

	for _, option := range options {
		err := option(&t)
		if err != nil {
			return nil, err
		}
	}
	return &t, nil
}

func (t *Termcorder) Start() error {
	err := t.cmd.Start()
	if err != nil {
		t.pty.Close()
		return err
	}
	defer t.pty.Close()
	defer t.restoreTermState()

	go func() {
		io.Copy(t.inputWriter, t.inputReader)
	}()

	io.Copy(t.outputWriter, t.outputReader)
	return nil
}

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
