package commands

import (
	"fmt"
	"net"
	"os"
)

type NetIn struct {
	Port uint16 `short:"p" long:"port" description:"container port" required:"true"`
}

func (command *NetIn) Execute(maybeHandle []string) error {
	container, err := globalClient().Lookup(handle(maybeHandle))
	failIf(err)

	hostPort, _, err := container.NetIn(0, uint32(command.Port))
	failIf(err)

	host := Globals.Target.Address
	if _, err := os.Stat(host); err == nil {
		host = "127.0.0.1"
	} else {
		host, _, err = net.SplitHostPort(host)
		failIf(err)
	}

	fmt.Println(net.JoinHostPort(host, fmt.Sprintf("%d", hostPort)))

	return nil
}
