package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/fatih/color"
	"github.com/jessevdk/go-flags"

	"github.com/cloudfoundry-incubator/garden"
	gclient "github.com/cloudfoundry-incubator/garden/client"
	gconn "github.com/cloudfoundry-incubator/garden/client/connection"

	"github.com/xoebus/gaol/commands"
)

func fail(err error) {
	fmt.Fprintln(os.Stderr, color.RedString("error:"), err)
	os.Exit(1)
}

func failIf(err error) {
	if err != nil {
		fail(err)
	}
}

func client(c *cli.Context) garden.Client {
	address := c.GlobalString("target")
	network := "tcp"
	if _, err := os.Stat(address); err == nil {
		network = "unix"
	}
	return gclient.New(gconn.New(network, address))
}

type command struct {
	name        string
	description string
	command     interface{}
}

func main() {
	parser := flags.NewParser(&commands.Globals, flags.HelpFlag|flags.PassDoubleDash)

	commands := []command{
		{"ping", "check to see if the garden host is up", &commands.Ping{}},
		{"create", "create a container", &commands.Create{}},
		{"destroy", "destroy a container", &commands.Destroy{}},
		{"list", "list running containers", &commands.List{}},
		{"run", "run a command in the container", &commands.Run{}},
		{"attach", "attach to a commmand running inside the container", &commands.Attach{}},
		{"shell", "open a shell in the container", &commands.Shell{}},
		{"stream-in", "stream data into the container", &commands.StreamIn{}},
		{"stream-out", "stream data out of the container", &commands.StreamOut{}},
		{"net-in", "map a port on the host to a port in the container", &commands.NetIn{}},
	}

	for _, command := range commands {
		_, err := parser.AddCommand(
			command.name,
			command.description,
			"",
			command.command,
		)
		failIf(err)
	}

	_, err := parser.Parse()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
