package commands

import (
	"fmt"
	"os"

	"github.com/cloudfoundry-incubator/garden"
	"github.com/mattn/go-shellwords"
)

type Run struct {
	Attach  bool     `short:"a" long:"attach" description:"attach to the process after it has started"`
	Dir     string   `short:"d" long:"dir" description:"current working directory of process"`
	User    string   `short:"u" long:"user" description:"user to run the process as"`
	Command string   `short:"g" long:"grace" description:"grace time (resetting ttl) of container"`
	Env     []string `short:"e" long:"env" description:"set environment variables"`
}

func (command *Run) Execute(maybeHandle []string) error {
	handle := handle(maybeHandle)
	container, err := globalClient().Lookup(handle)
	failIf(err)

	var processIo garden.ProcessIO
	if command.Attach {
		processIo = garden.ProcessIO{
			Stdin:  os.Stdin,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		}
	} else {
		processIo = garden.ProcessIO{}
	}

	args, err := shellwords.Parse(command.Command)
	failIf(err)

	process, err := container.Run(garden.ProcessSpec{
		Path: args[0],
		Args: args[1:],
		Dir:  command.Dir,
		User: command.User,
		Env:  command.Env,
	}, processIo)
	failIf(err)

	if command.Attach {
		status, err := process.Wait()
		failIf(err)
		os.Exit(status)
	} else {
		fmt.Println(process.ID())
	}

	return nil
}
