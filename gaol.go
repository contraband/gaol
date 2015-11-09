package main

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"

	"github.com/contraband/gaol/commands"
)

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

		if err != nil {
			panic(err)
		}
	}

	_, err := parser.Parse()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
