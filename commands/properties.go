package commands

import (
	"encoding/json"
	"fmt"
	"os"
)

type Properties struct {
	AsJSON bool `short:"j" long:"json" description:"format output as json"`
}

func (command *Properties) Execute(maybeHandle []string) error {
	container, err := globalClient().Lookup(handle(maybeHandle))
	failIf(err)

	properties, err := container.Properties()
	failIf(err)

	memoryLimits, err := container.CurrentMemoryLimits()
	failIf(err)

	if command.AsJSON {
		toPrint, err := json.MarshalIndent(properties, "", "  ")
		failIf(err)
		os.Stdout.Write(toPrint)
	} else {
		for k, v := range properties {
			fmt.Printf("%s\t%s\n", k, v)
		}
		fmt.Printf("memory.limit\t%d bytes\n", memoryLimits.LimitInBytes)
	}

	return nil
}
