package commands

import (
	"github.com/cloudfoundry-incubator/garden"
	gclient "github.com/cloudfoundry-incubator/garden/client"
	gconn "github.com/cloudfoundry-incubator/garden/client/connection"
)

func globalClient() garden.Client {
	network := Globals.Target.Network
	address := Globals.Target.Address

	return gclient.New(gconn.New(network, address))
}
