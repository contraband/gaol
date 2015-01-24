package main

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/codegangsta/cli"
	"github.com/pivotal-golang/archiver/compressor"

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
	target := c.GlobalString("target")
	return gclient.New(gconn.New("tcp", target))
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
			Usage:  "server to which commands are sent",
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
			Action: func(c *cli.Context) {
				container, err := client(c).Create(garden.ContainerSpec{})
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
			Action: func(c *cli.Context) {
				containers, err := client(c).Containers(nil)
				failIf(err)

				for _, container := range containers {
					fmt.Println(container.Handle())
				}
			},
		},
		{
			Name:  "stream-in",
			Usage: "stream data into the container",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "to-file, t",
					Usage: "destination path in the container",
				},
			},
			Action: func(c *cli.Context) {
				handle := handle(c)
				dst := c.String("to-file")

				container, err := client(c).Lookup(handle)
				failIf(err)

				// perform dance to get correct file names
				tmpDir, err := ioutil.TempDir("", "gaol")
				failIf(err)
				defer os.RemoveAll(tmpDir)

				tmp, err := os.Create(filepath.Join(tmpDir, filepath.Base(dst)))
				failIf(err)

				_, err = io.Copy(tmp, os.Stdin)
				failIf(err)

				err = tmp.Close()
				failIf(err)

				reader, writer := io.Pipe()
				go func(w io.WriteCloser) {
					err := compressor.WriteTar(tmp.Name(), w)
					failIf(err)
					w.Close()
				}(writer)

				err = container.StreamIn(filepath.Dir(dst), reader)
				failIf(err)
			},
		},
		{
			Name:  "stream-out",
			Usage: "stream data out of the container",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "from-file, f",
					Usage: "source path in the container",
				},
			},
			Action: func(c *cli.Context) {
				handle := handle(c)
				src := c.String("from-file")

				container, err := client(c).Lookup(handle)
				failIf(err)

				output, err := container.StreamOut(src)
				failIf(err)

				tr := tar.NewReader(output)
				_, err = tr.Next()
				failIf(err)

				_, err = io.Copy(os.Stdout, tr)
				failIf(err)
			},
		},
	}

	app.Run(os.Args)
}
