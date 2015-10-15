package commands

import (
	"os"

	"github.com/cloudfoundry-incubator/garden"
)

type StreamIn struct {
	Destination string `short:"d" long:"destination" description:"destination path in the container" required:"true"`
}

func (command *StreamIn) Execute(maybeHandle []string) error {
	container, err := globalClient().Lookup(handle(maybeHandle))
	failIf(err)

	streamInSpec := garden.StreamInSpec{
		Path:      command.Destination,
		TarStream: os.Stdin,
	}

	err = container.StreamIn(streamInSpec)
	failIf(err)

	return nil
}
