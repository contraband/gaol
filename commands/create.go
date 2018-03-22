package commands

import (
	"fmt"
	"strings"
	"time"

	"code.cloudfoundry.org/garden"
)

type Create struct {
	Handle     string        `short:"n" long:"handle" description:"name to give container"`
	RootFS     string        `short:"r" long:"rootfs" description:"rootfs image with which to create the container"`
	Username   string        `long:"username" description:"username for private registry"`
	Password   string        `long:"password" description:"password for private registry"`
	Env        []string      `short:"e" long:"env" description:"set environment variables"`
	Grace      time.Duration `short:"g" long:"grace" description:"grace time (resetting ttl) of container"`
	Privileged bool          `short:"p" long:"privileged" description:"privileged user in the container is privileged in the host"`
	Network    string        `long:"network" description:"the subnet of the container"`
	BindMounts []string      `short:"m" long:"bind-mount" description:"bind mount host-path:container-path:optional-bind-mount-mode"`
}

func (command *Create) Execute(args []string) error {
	var bindMounts []garden.BindMount

	for _, pair := range command.BindMounts {
		segs := strings.SplitN(pair, ":", 3)
		if len(segs) < 2 {
			fail(fmt.Errorf("invalid bind-mount segment (must be host-path:container-path:optional-bind-mount-option): %s", pair))
		}

		var bindMountMode garden.BindMountMode
		var err error
		if bindMountMode, err = getBindMountMode(segs); err != nil {
			fail(err)
		}

		bindMounts = append(bindMounts, garden.BindMount{
			SrcPath: segs[0],
			DstPath: segs[1],
			Mode:    bindMountMode,
			Origin:  garden.BindMountOriginHost,
		})
	}

	container, err := globalClient().Create(garden.ContainerSpec{
		Handle:     command.Handle,
		GraceTime:  command.Grace,
		Image:      garden.ImageRef{URI: command.RootFS, Username: command.Username, Password: command.Password},
		Privileged: command.Privileged,
		Env:        command.Env,
		Network:    command.Network,
		BindMounts: bindMounts,
	})

	failIf(err)

	fmt.Println(container.Handle())

	return nil
}

func getBindMountMode(bindMountSegments []string) (garden.BindMountMode, error) {
	if len(bindMountSegments) < 3 {
		return garden.BindMountModeRW, nil
	}

	switch bindMountSegments[2] {
	case "ro":
		return garden.BindMountModeRO, nil
	case "rw":
		return garden.BindMountModeRW, nil
	default:
		return 0, fmt.Errorf("invalid bind mount mode (must be either 'rw' (read-write), or 'ro' (read-only))")
	}
}
