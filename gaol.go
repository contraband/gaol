package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/codegangsta/cli"
	"github.com/kr/pty"
	"github.com/mattn/go-shellwords"
	"github.com/pkg/term"

	"github.com/cloudfoundry-incubator/garden"
	gclient "github.com/cloudfoundry-incubator/garden/client"
	gconn "github.com/cloudfoundry-incubator/garden/client/connection"
)

func fail(err error) {
	fmt.Fprintln(os.Stderr, "failed:", err)
	os.Exit(1)
}

func failIf(err error) {
	if err != nil {
		fail(err)
	}
}

func client(c *cli.Context) garden.Client {
	address := c.GlobalString("target")
	network := "tcp"
	if _, err := os.Stat(address); err == nil {
		network = "unix"
	}
	return gclient.New(gconn.New(network, address))
}

func handle(c *cli.Context) string {
	if len(c.Args()) == 0 {
		fail(errors.New("must provide container handle"))
	}
	return c.Args().First()
}

func main() {
	app := cli.NewApp()
	app.Name = "gaol"
	app.Usage = "a cli for garden"
	app.Version = "0.0.1"
	app.Author = "Chris Brown"
	app.Email = "cbrown@pivotal.io"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "target, t",
			Value:  "localhost:7777",
			Usage:  "server or unix socket to which commands are sent",
			EnvVar: "GAOL_TARGET",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:  "ping",
			Usage: "check if the server is running",
			Action: func(c *cli.Context) {
				err := client(c).Ping()
				failIf(err)
			},
		},
		{
			Name:  "create",
			Usage: "create a container",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "handle, n",
					Usage: "name to give container",
				},
				cli.StringFlag{
					Name:  "rootfs, r",
					Usage: "rootfs image with which to create the container",
				},
				cli.StringSliceFlag{
					Name:  "env, e",
					Usage: "set environment variables",
					Value: &cli.StringSlice{},
				},
				cli.DurationFlag{
					Name:  "grace, g",
					Usage: "grace time (resetting ttl) of container",
					Value: 5 * time.Minute,
				},
				cli.BoolFlag{
					Name:  "privileged, p",
					Usage: "privileged user in container is privileged in host",
				},
				cli.StringFlag{
					Name:  "network",
					Usage: "the subnet of the container",
				},
				cli.StringSliceFlag{
					Name:  "bind-mount, m",
					Usage: "bind-mount host-path:container-path",
					Value: &cli.StringSlice{},
				},
			},
			Action: func(c *cli.Context) {
				handle := c.String("handle")
				grace := c.Duration("grace")
				rootfs := c.String("rootfs")
				env := c.StringSlice("env")
				privileged := c.Bool("privileged")
				network := c.String("network")
				mounts := c.StringSlice("bind-mount")

				var bindMounts []garden.BindMount
				for _, pair := range mounts {
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

				container, err := client(c).Create(garden.ContainerSpec{
					Handle:     handle,
					GraceTime:  grace,
					RootFSPath: rootfs,
					Privileged: privileged,
					Env:        env,
					Network:    network,
					BindMounts: bindMounts,
				})
				failIf(err)

				fmt.Println(container.Handle())
			},
		},
		{
			Name:  "destroy",
			Usage: "destroy a container",
			Action: func(c *cli.Context) {
				client := client(c)
				handles := c.Args()

				for _, handle := range handles {
					err := client.Destroy(handle)
					failIf(err)
				}
			},
		},
		{
			Name:  "list",
			Usage: "get a list of running containers",
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:  "properties, p",
					Usage: "filter by properties (name=val)",
					Value: &cli.StringSlice{},
				},
				cli.BoolFlag{
					Name:  "verbose, v",
					Usage: "print additional details about each container",
				},
				cli.StringFlag{
					Name:  "separator",
					Usage: "separator to print between containers in verbose mode",
					Value: "\n",
				},
			},
			Action: func(c *cli.Context) {
				separator := c.String("separator")

				properties := garden.Properties{}
				for _, prop := range c.StringSlice("properties") {
					segs := strings.SplitN(prop, "=", 2)
					if len(segs) < 2 {
						fail(errors.New("malformed property pair (must be name=value)"))
					}

					properties[segs[0]] = segs[1]
				}

				containers, err := client(c).Containers(properties)
				failIf(err)

				verbose := c.Bool("verbose")

				for _, container := range containers {
					fmt.Println(container.Handle())

					if verbose {
						props, _ := container.Properties()
						for k, v := range props {
							fmt.Printf("  %s=%s\n", k, v)
						}

						fmt.Print(separator)
					}
				}
			},
		},
		{
			Name:  "run",
			Usage: "run a command in a container",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "attach, a",
					Usage: "attach to the process after it is started",
				},
				cli.StringFlag{
					Name:  "dir, d",
					Usage: "current working directory of process",
				},
				cli.StringFlag{
					Name:  "user, u",
					Usage: "user to run the process as",
				},
				cli.StringFlag{
					Name:  "command, c",
					Usage: "the command to run",
				},
				cli.StringSliceFlag{
					Name:  "env, e",
					Usage: "set environment variables",
					Value: &cli.StringSlice{},
				},
			},
			Action: func(c *cli.Context) {
				attach := c.Bool("attach")
				dir := c.String("dir")
				user := c.String("user")
				command := c.String("command")
				env := c.StringSlice("env")

				handle := handle(c)
				container, err := client(c).Lookup(handle)
				failIf(err)

				var processIo garden.ProcessIO
				if attach {
					processIo = garden.ProcessIO{
						Stdin:  os.Stdin,
						Stdout: os.Stdout,
						Stderr: os.Stderr,
					}
				} else {
					processIo = garden.ProcessIO{}
				}

				args, err := shellwords.Parse(command)
				failIf(err)

				process, err := container.Run(garden.ProcessSpec{
					Path: args[0],
					Args: args[1:],
					Dir:  dir,
					User: user,
					Env:  env,
				}, processIo)
				failIf(err)

				if attach {
					status, err := process.Wait()
					failIf(err)
					os.Exit(status)
				} else {
					fmt.Println(process.ID())
				}
			},
		},
		{
			Name:  "attach",
			Usage: "attach to command running in the container",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "pid, p",
					Usage: "process id to connect to",
				},
			},
			Action: func(c *cli.Context) {
				pid := c.String("pid")
				if pid == "" {
					err := errors.New("must specify pid to attach to")
					failIf(err)
				}

				handle := handle(c)
				container, err := client(c).Lookup(handle)
				failIf(err)

				process, err := container.Attach(pid, garden.ProcessIO{
					Stdin:  os.Stdin,
					Stdout: os.Stdout,
					Stderr: os.Stderr,
				})
				failIf(err)

				_, err = process.Wait()
				failIf(err)
			},
		},
		{
			Name:  "shell",
			Usage: "open a shell inside the running container",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "user, u",
					Usage: "user to open shell as",
				},
			},
			Action: func(c *cli.Context) {
				container, err := client(c).Lookup(handle(c))
				failIf(err)

				term, err := term.Open(os.Stdin.Name())
				failIf(err)

				err = term.SetRaw()
				failIf(err)

				rows, cols, err := pty.Getsize(os.Stdin)
				failIf(err)

				process, err := container.Run(garden.ProcessSpec{
					User: c.String("user"),
					Path: "/bin/sh",
					Args: []string{"-l"},
					Env:  []string{"TERM=" + os.Getenv("TERM")},
					TTY: &garden.TTYSpec{
						WindowSize: &garden.WindowSize{
							Rows:    rows,
							Columns: cols,
						},
					},
				}, garden.ProcessIO{
					Stdin:  term,
					Stdout: term,
					Stderr: term,
				})
				if err != nil {
					term.Restore()
					failIf(err)
				}

				resized := make(chan os.Signal, 10)
				signal.Notify(resized, syscall.SIGWINCH)

				go func() {
					for {
						<-resized

						rows, cols, err := pty.Getsize(os.Stdin)
						if err == nil {
							process.SetTTY(garden.TTYSpec{
								WindowSize: &garden.WindowSize{
									Rows:    rows,
									Columns: cols,
								},
							})
						}
					}
				}()

				process.Wait()
				term.Restore()
			},
		},
		{
			Name:  "stream-in",
			Usage: "stream data into the container",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "destination, d",
					Usage: "destination path in the container",
				},
			},
			Action: func(c *cli.Context) {
				handle := handle(c)

				dst := c.String("destination")
				if dst == "" {
					fail(errors.New("missing --destination flag"))
				}

				container, err := client(c).Lookup(handle)
				failIf(err)

				streamInSpec := garden.StreamInSpec{
					Path:      dst,
					TarStream: os.Stdin,
				}

				err = container.StreamIn(streamInSpec)
				failIf(err)
			},
		},
		{
			Name:  "stream-out",
			Usage: "stream data out of the container",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "source, s",
					Usage: "source path in the container",
				},
			},
			Action: func(c *cli.Context) {
				handle := handle(c)

				src := c.String("source")
				if src == "" {
					fail(errors.New("missing --source flag"))
				}

				container, err := client(c).Lookup(handle)
				failIf(err)

				streamOutSpec := garden.StreamOutSpec{Path: src}
				output, err := container.StreamOut(streamOutSpec)
				failIf(err)

				io.Copy(os.Stdout, output)
			},
		},
		{
			Name:  "net-in",
			Usage: "map a port on the host to a port in the container",
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "port, p",
					Usage: "container port",
				},
			},
			Action: func(c *cli.Context) {
				target := c.GlobalString("target")
				requestedContainerPort := uint32(c.Int("port"))

				if target == "" {
					fail(errors.New("target must be set"))
				}

				handle := handle(c)
				container, err := client(c).Lookup(handle)
				failIf(err)

				hostPort, _, err := container.NetIn(0, requestedContainerPort)
				failIf(err)

				host, _, err := net.SplitHostPort(target)
				failIf(err)

				fmt.Println(net.JoinHostPort(host, fmt.Sprintf("%d", hostPort)))
			},
		},
	}

	app.Run(os.Args)
}
