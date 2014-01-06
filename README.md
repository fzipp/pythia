Pythia is a browser based user interface for the Go source code oracle,
which is a source code comprehension tool for Go programs.

For more information on the Go oracle, see the [original announcement](https://groups.google.com/d/msg/golang-nuts/CwdIJZs6Tfc/GX7ixTK_Dd4J).

![Screenshot](https://raw.github.com/fzipp/pythia/gh-pages/images/pythia_screenshot.png)


Installing from source
----------------------

Building Pythia requires at least Go version 1.2 or higher.

To install, run

    $ go get github.com/fzipp/pythia

You will now find a `pythia` binary in your `$GOPATH/bin` directory.

Usage
-----

Start the web application with a package path, e.g.:

    $ pythia net/http

By default it will listen on port :8080 and try to launch the application
in your browser. You can choose a different port via the `-http` flag, e.g.:

    $ pythia -http :6060 fmt

Run `pythia -help` for more information.

