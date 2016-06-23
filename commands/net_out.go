package commands

import (
	"errors"
	"fmt"
	"net"

	"github.com/cloudfoundry-incubator/garden"
)

type NetOut struct {
	Protocol  string `short:"p" required:"true" long:"protocol" choice:"tcp" choice:"udp" description:"protocol to whitelist, only supports tcp or udp"`
	StartIP   IPFlag `long:"ip-start" required:"true" description:"start of IP range to whitelist, inclusive"`
	EndIP     IPFlag `long:"ip-end" required:"true" description:"end of IP range to whitelist, inclusive"`
	StartPort uint16 `long:"port-start" required:"true" description:"start of port range to whitelist, inclusive"`
	EndPort   uint16 `long:"port-end" required:"true" description:"end of port range to whitelist, inclusive"`
}

func (command *NetOut) Execute(maybeHandle []string) error {
	container, err := globalClient().Lookup(handle(maybeHandle))
	failIf(err)

	var protocol garden.Protocol
	switch command.Protocol {
	case "tcp":
		protocol = garden.ProtocolTCP
	case "udp":
		protocol = garden.ProtocolUDP
	default:
		fail(errors.New("unrecognised protocol"))
	}

	ipRange := garden.IPRange{
		Start: command.StartIP.IP(),
		End:   command.EndIP.IP(),
	}

	portRange := garden.PortRange{
		Start: command.StartPort,
		End:   command.EndPort,
	}

	netOutRule := garden.NetOutRule{
		Protocol: protocol,
		Networks: []garden.IPRange{ipRange},
		Ports:    []garden.PortRange{portRange},
	}

	err = container.NetOut(netOutRule)
	failIf(err)

	fmt.Println("applied")

	return nil
}

type IPFlag net.IP

func (f *IPFlag) UnmarshalFlag(value string) error {
	parsedIP := net.ParseIP(value)
	if parsedIP == nil {
		return fmt.Errorf("invalid IP: '%s'", value)
	}

	*f = IPFlag(parsedIP)

	return nil
}

func (f IPFlag) IP() net.IP {
	return net.IP(f)
}
