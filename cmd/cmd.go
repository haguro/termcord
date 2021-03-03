package cmd

import (
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

func Run(fname, cmd string) error {
	var err error

	f, err := os.OpenFile(fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	c := exec.Command(cmd)

	ptmx, err := pty.Start(c)
	if err != nil {
		return err
	}
	defer ptmx.Close()

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

	inputMWriter := io.MultiWriter(ptmx, f)
	outputMWriter := io.MultiWriter(os.Stdout, f) // do we needs to add os.Stderr here too?

	go func() {
		io.Copy(inputMWriter, os.Stdin)
	}()

	go func() {
		//We can send anything to the ptmx with `Write`, `WriteString`..etc
		time.Sleep(3 * time.Second)
		ptmx.Write([]byte{4})
	}()

	io.Copy(outputMWriter, ptmx)

	return nil
}
