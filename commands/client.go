package commands

import (
	"code.cloudfoundry.org/garden"
	gclient "code.cloudfoundry.org/garden/client"
	gconn "code.cloudfoundry.org/garden/client/connection"
)

func globalClient() garden.Client {
	network := Globals.Target.Network
	address := Globals.Target.Address

	return gclient.New(gconn.New(network, address))
}
