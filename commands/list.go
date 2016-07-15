package commands

import (
	"errors"
	"fmt"
	"strings"

	"code.cloudfoundry.org/garden"
)

type List struct {
	Properties []string `short:"p" long:"properties" description:"filter by properties (name=value)"`
	Verbose    bool     `short:"v" long:"verbose" description:"print additional details about each container"`
	Separator  string   `long:"separator" description:"separator to print between containers in verbose mode"`
}

func (command *List) Execute(args []string) error {
	properties := garden.Properties{}
	for _, prop := range command.Properties {
		segs := strings.SplitN(prop, "=", 2)
		if len(segs) < 2 {
			fail(errors.New("malformed property pair (must be name=value)"))
		}

		properties[segs[0]] = segs[1]
	}

	containers, err := globalClient().Containers(properties)
	failIf(err)

	for _, container := range containers {
		fmt.Println(container.Handle())

		if command.Verbose {
			props, _ := container.Properties()
			for k, v := range props {
				fmt.Printf("  %s=%s\n", k, v)
			}

			fmt.Print(command.Separator)
		}
	}

	return nil
}
