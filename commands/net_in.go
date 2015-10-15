package commands

import (
	"fmt"
	"net"
)

type NetIn struct {
	Port uint16 `short:"p" long:"port" description:"container port" required:"true"`
}

func (command *NetIn) Execute(maybeHandle []string) error {
	container, err := globalClient().Lookup(handle(maybeHandle))
	failIf(err)

	hostPort, _, err := container.NetIn(0, uint32(command.Port))
	failIf(err)

	host, _, err := net.SplitHostPort(Globals.Target.Address)
	failIf(err)

	fmt.Println(net.JoinHostPort(host, fmt.Sprintf("%d", hostPort)))

	return nil
}
