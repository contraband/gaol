package commands

import (
	"os"

	"code.cloudfoundry.org/garden"
)

type Attach struct {
	Process string `short:"p" long:"pid" description:"process id to connect to" required:"true"`
}

func (command *Attach) Execute(maybeHandle []string) error {
	handle := handle(maybeHandle)
	container, err := globalClient().Lookup(handle)
	failIf(err)

	process, err := container.Attach(command.Process, garden.ProcessIO{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
	failIf(err)

	_, err = process.Wait()
	failIf(err)

	return nil
}
