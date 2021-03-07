package cmd

import (
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
	File    io.Writer
	Cmd     *exec.Cmd
	Iactive bool
}

//Run creates the script file, creates a new pty and runs the command in that pty
func Run(c Config) error {
	var err error

	ptmx, err := pty.Start(c.Cmd)
	if err != nil {
		return err
	}
	defer ptmx.Close()

	if c.Iactive { //if interactive
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
			panic(err)
		}
		defer term.Restore(int(os.Stdin.Fd()), oldState)
	}
	//inputMWriter := io.MultiWriter(ptmx, f) //TODO: Add option to record stdin as well (as with script -k)
	outputMWriter := io.MultiWriter(os.Stdout, c.File)

	go func() {
		io.Copy(ptmx, os.Stdin)
	}()

	io.Copy(outputMWriter, ptmx)

	return c.Cmd.Wait()
}
