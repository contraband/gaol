package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/codegangsta/cli"
	"github.com/fatih/color"
	"github.com/jessevdk/go-flags"
	"github.com/kr/pty"
	"github.com/pkg/term"

	"github.com/cloudfoundry-incubator/garden"
	gclient "github.com/cloudfoundry-incubator/garden/client"
	gconn "github.com/cloudfoundry-incubator/garden/client/connection"

	"github.com/xoebus/gaol/commands"
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

func client(c *cli.Context) garden.Client {
	address := c.GlobalString("target")
	network := "tcp"
	if _, err := os.Stat(address); err == nil {
		network = "unix"
	}
	return gclient.New(gconn.New(network, address))
}

type command struct {
	name        string
	description string
	command     interface{}
}

func main() {
	parser := flags.NewParser(&commands.Globals, flags.HelpFlag|flags.PassDoubleDash)

	commands := []command{
		{"ping", "check to see if the garden host is up", &commands.Ping{}},
		{"create", "create a container", &commands.Create{}},
		{"destroy", "destroy a container", &commands.Destroy{}},
		{"list", "list running containers", &commands.List{}},
		{"run", "run a command in the container", &commands.Run{}},
	}

	for _, command := range commands {
		_, err := parser.AddCommand(
			command.name,
			command.description,
			"",
			command.command,
		)
		failIf(err)
	}

	_, err := parser.Parse()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(0)

	app := cli.NewApp()
	app.Name = "gaol"
	app.Usage = "a cli for garden"
	app.Version = "0.0.1"
	app.Author = "Chris Brown"
	app.Email = "cbrown@pivotal.io"

	app.Commands = []cli.Command{
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
