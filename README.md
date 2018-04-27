Pythia is a browser based user interface for the Go source code guru,
which is a tool for navigating Go code.

For more information on the Go guru, see [Using Go Guru](http://golang.org/s/using-guru).

![Screenshot](https://raw.github.com/fzipp/pythia/gh-pages/images/pythia_screenshot.png)


Installing from source
----------------------

Building Pythia requires `vgo`:

    $ go get golang.org/x/vgo

To install, run

    $ git clone https://github.com/fzipp/pythia
    $ cd pythia
    $ vgo install

You will now find a `pythia` binary in your `$GOPATH/bin` directory.

Running Pythia also requires `guru`:

    $ go get golang.org/x/tools/cmd/guru

Usage
-----

Start the web application with a package path, e.g.:

    $ pythia net/http

By default it will listen on port :8080 and try to launch the application
in your browser. You can choose a different port via the `-http` flag, e.g.:

    $ pythia -http :6060 fmt

Run `pythia -help` for more information.

