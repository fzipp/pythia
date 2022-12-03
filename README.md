# Pythia

![Build Status](https://github.com/fzipp/pythia/workflows/build/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/fzipp/pythia)](https://goreportcard.com/report/github.com/fzipp/pythia)

Pythia is a browser based user interface for the Go source code guru,
which is a tool for navigating Go code.

For more information on the Go guru, see [Using Go Guru](https://go.dev/s/using-guru).

![Screenshot](https://raw.github.com/fzipp/pythia/gh-pages/images/pythia_screenshot.png)

## Installing from source

To install, run

    $ go install github.com/fzipp/pythia@latest

You will now find a `pythia` binary in your `$GOBIN` or `$GOPATH/bin` directory.

Running Pythia also requires `guru`:

    $ go install golang.org/x/tools/cmd/guru@latest

## Usage

Start the web application with a package path, e.g.:

    $ pythia net/http

By default it will listen on port :8080 and try to launch the application
in your browser. You can choose a different port via the `-http` flag, e.g.:

    $ pythia -http :6060 fmt

Run `pythia -help` for more information.
