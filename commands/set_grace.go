package commands

import (
	"time"
)

type SetGraceTime struct {
	Grace time.Duration `short:"g" long:"grace" description:"grace time (resetting ttl) of container" required:"true"`
}

func (command *SetGraceTime) Execute(maybeHandle []string) error {
	container, err := globalClient().Lookup(handle(maybeHandle))
	failIf(err)

	err = container.SetGraceTime(command.Grace)
	failIf(err)
	return nil
}
