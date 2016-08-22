<img src="https://cdn.rawgit.com/contraband/gaol/master/docs/images/gaol.svg" align="left" width="192px" height="192px"/>

# gaol

> *A CLI for Garden (pronounced 'jail')*

[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](/LICENSE)

<br />

## about

Gaol is a simple, fast, and composable CLI for the Garden container server. It
is designed to be used with other command line tools. It is suitable for use in
scripts and other headless operations.


## installation

Download the latest binary from the [releases page][releases], make it
executable, and put it somewhere in your path.

[releases]: https://github.com/contraband/gaol/releases


## usage

All of the commands are documented in the help and usage section of the CLI.
You can see this by running:

    gaol --help

Some of the more common commands are:

    # creating a container
    $ gaol create
    conabc123

    # run a command inside a container
    $ gaol run conabc123 --attach --command "date"
    Sat  7 Feb 2015 15:14:17 GMT

    # run a process in the background and then attach to it
    $ gaol run conabc123 --command 'sh -c "while true; do date; sleep 1; done"'
    5
    $ gaol attach conabc123 --pid 5
    Sat  7 Feb 2015 15:14:45 GMT
    Sat  7 Feb 2015 15:14:46 GMT
    Sat  7 Feb 2015 15:14:47 GMT

    # open a shell inside a new container
    $ gaol shell $(gaol create)

    # copying files into a container
    $ tar c file.txt | gaol stream-in conabc123 --destination /etc/file.txt

    # copying files from one container to another
    $ gaol stream-out abc -s ./foo | gaol stream-in def -d ./foo

    # destroy all containers
    $ gaol list | xargs gaol destroy


## links

Garden
https://github.com/cloudfoundry/garden

Gaol Definition
http://en.wiktionary.org/wiki/gaol
