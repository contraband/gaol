package commands

import (
	"errors"
	"os"
)

type GlobalOptions struct {
	Target TargetFlag `short:"t" long:"target" description:"garden endpoint" default:"127.0.0.1:7777" env:"GAOL_TARGET"`
}

var Globals GlobalOptions

type TargetFlag struct {
	Network string
	Address string
}

func (target *TargetFlag) UnmarshalFlag(address string) error {
	network := "tcp"
	if _, err := os.Stat(address); err == nil {
		network = "unix"
	}

	target.Address = address
	target.Network = network

	return nil
}

func handle(args []string) string {
	if len(args) == 0 {
		fail(errors.New("must provide container handle"))
	}
	return args[0]
}
