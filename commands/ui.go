package commands

import (
	"fmt"
	"os"

	"github.com/fatih/color"
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
