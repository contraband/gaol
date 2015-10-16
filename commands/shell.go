package commands

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/cloudfoundry-incubator/garden"
	"github.com/kr/pty"
	"github.com/pkg/term"
)

type Shell struct {
	User string `short:"u" long:"user" description:"user to open shell as" default:"root"`
}

func (command *Shell) Execute(maybeHandle []string) error {
	container, err := globalClient().Lookup(handle(maybeHandle))
	failIf(err)

	term, err := term.Open(os.Stdin.Name())
	failIf(err)

	err = term.SetRaw()
	failIf(err)

	rows, cols, err := pty.Getsize(os.Stdin)
	failIf(err)

	process, err := container.Run(garden.ProcessSpec{
		User: command.User,
		Path: "/bin/sh",
		Args: []string{"-l"},
		Env:  []string{"TERM=" + os.Getenv("TERM")},
		TTY: &garden.TTYSpec{
			WindowSize: &garden.WindowSize{
				Rows:    rows,
				Columns: cols,
			},
		},
	}, garden.ProcessIO{
		Stdin:  term,
		Stdout: term,
		Stderr: term,
	})
	if err != nil {
		term.Restore()
		failIf(err)
	}

	resized := make(chan os.Signal, 10)
	signal.Notify(resized, syscall.SIGWINCH)

	go func() {
		for {
			<-resized

			rows, cols, err := pty.Getsize(os.Stdin)
			if err == nil {
				process.SetTTY(garden.TTYSpec{
					WindowSize: &garden.WindowSize{
						Rows:    rows,
						Columns: cols,
					},
				})
			}
		}
	}()

	process.Wait()
	term.Restore()

	return nil
}
