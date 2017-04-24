package commands

import (
	"fmt"
	"strings"
	"time"

	"code.cloudfoundry.org/garden"
)

type Create struct {
	Handle      string        `short:"n" long:"handle" description:"name to give container"`
	RootFS      string        `short:"r" long:"rootfs" description:"rootfs image with which to create the container"`
	Env         []string      `short:"e" long:"env" description:"set environment variables"`
	Grace       time.Duration `short:"g" long:"grace" description:"grace time (resetting ttl) of container"`
	Privileged  bool          `short:"p" long:"privileged" description:"privileged user in the container is privileged in the host"`
	Network     string        `long:"network" description:"the subnet of the container"`
	BindMounts  []string      `short:"m" long:"bind-mount" description:"bind mount host-path:container-path"`
	LimitMemory uint64        `short:"k" long:"limit-memory" description:"limits the memory used by the container. Value in bytes"`
}

func (command *Create) Execute(args []string) error {
	var bindMounts []garden.BindMount

	for _, pair := range command.BindMounts {
		segs := strings.SplitN(pair, ":", 2)
		if len(segs) != 2 {
			fail(fmt.Errorf("invalid bind-mount segment (must be host-path:container-path): %s", pair))
		}

		bindMounts = append(bindMounts, garden.BindMount{
			SrcPath: segs[0],
			DstPath: segs[1],
			Mode:    garden.BindMountModeRW,
			Origin:  garden.BindMountOriginHost,
		})
	}

	container, err := globalClient().Create(garden.ContainerSpec{
		Handle:     command.Handle,
		GraceTime:  command.Grace,
		RootFSPath: command.RootFS,
		Privileged: command.Privileged,
		Env:        command.Env,
		Network:    command.Network,
		BindMounts: bindMounts,
		Limits: garden.Limits{
			Memory: garden.MemoryLimits{
				LimitInBytes: command.LimitMemory,
			},
		},
	})

	failIf(err)

	fmt.Println(container.Handle())

	return nil
}
