// Copyright 2013 Frederik Zipp.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Pythia is a web application front-end for the Go source code guru.
package main

import (
	"cmp"
	"flag"
	"fmt"
	"go/build"
	"go/token"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"slices"
	"strings"

	"golang.org/x/tools/go/loader"
)

var (
	guruPath = ""
	httpAddr = flag.String("http", ":8080", "HTTP listen address")
	verbose  = flag.Bool("v", false, "Verbose mode: print incoming queries")
	open     = flag.Bool("open", true, "Try to open browser")
	tags     = flag.String("tags", "", "Tags to use when importing packages")
	args     []string
	files    []string
	packages []*loader.PackageInfo
	prog     *loader.Program
	scope    string
)

const useHelp = "Run 'pythia -help' for more information.\n"

const helpMessage = `Web frontend for the Go source code guru.
Usage: pythia [<flag> ...] <args> ...

The -http flag specifies the HTTP service address (e.g., ':6060').

The -tags flag specifies comma separated tags to use when importing
code (e.g., 'foo,!darwin').

The -open flag determines, whether the application should try to
open the browser. It is set to 'true' by default. If set to 'false'
the browser will not be launched.

The -v flag enables verbose mode, in which every incoming query
to the guru is logged to the standard output.
` + loader.FromArgsUsage + `
Examples:

Start pythia with the scope of package guru:
% pythia golang.org/x/tools/cmd/guru

Start pythia with the scope of package image/png on port 8081,
but don't open the browser:
% pythia -http=:8081 -open=false image/png

`

func main() {
	var err error
	// Check if guru is in the path.
	guruPath, err = exec.LookPath("guru")
	if err != nil {
		log.Fatal("Can't find guru in your path")
		return
	}
	flag.Usage = func() {}
	flag.CommandLine.Init(os.Args[0], flag.ContinueOnError)
	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			fmt.Print(helpMessage)
		} else {
			fmt.Fprint(os.Stderr, useHelp)
		}
		os.Exit(2)
	}
	args = flag.Args()
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, "Error: no package arguments.\n"+useHelp)
		os.Exit(2)
	}

	settings := build.Default
	settings.BuildTags = strings.Split(*tags, ",")
	conf := loader.Config{Build: &settings}
	_, err = conf.FromArgs(args, true)
	exitOn(err)
	prog, err = conf.Load()
	exitOn(err)
	files = scopeFiles(prog)
	packages = sortedPackages(prog)
	scope = strings.Join(args, ",")

	registerHandlers()

	srv := &http.Server{Addr: *httpAddr}
	l, err := net.Listen("tcp", srv.Addr)
	exitOn(err)
	if *open {
		url := fmt.Sprintf("http://localhost%s/", *httpAddr)
		if !startBrowser(url) {
			fmt.Println(url)
		}
	}
	exitError(srv.Serve(l))
}

func registerHandlers() {
	http.HandleFunc("GET /", serveIndex)
	http.HandleFunc("GET /source", serveSource)
	http.HandleFunc("GET /file", serveFile)
	http.HandleFunc("GET /query", serveQuery)
	http.Handle("GET /static/", http.FileServer(http.FS(staticFiles)))
}

// sortedPackages returns all packages of a program, sorted by package path.
func sortedPackages(prog *loader.Program) []*loader.PackageInfo {
	pkgs := make([]*loader.PackageInfo, 0, len(prog.AllPackages))
	for _, p := range prog.AllPackages {
		pkgs = append(pkgs, p)
	}
	slices.SortFunc(pkgs, func(a, b *loader.PackageInfo) int {
		return cmp.Compare(a.Pkg.Path(), b.Pkg.Path())
	})
	return pkgs
}

// scopeFiles returns a new slice containing the full paths of all the files
// imported by the loader, sorted in increasing order.
func scopeFiles(prog *loader.Program) []string {
	var files []string
	prog.Fset.Iterate(func(f *token.File) bool {
		files = append(files, f.Name())
		return true
	})
	slices.Sort(files)
	return files
}

// startBrowser tries to open the URL in a browser
// and reports whether it succeeds.
func startBrowser(url string) bool {
	// try to start the browser
	var args []string
	switch runtime.GOOS {
	case "darwin":
		args = []string{"open"}
	case "windows":
		args = []string{"cmd", "/c", "start"}
	default:
		args = []string{"xdg-open"}
	}
	cmd := exec.Command(args[0], append(args[1:], url)...)
	return cmd.Start() == nil
}

func exitOn(err error) {
	if err != nil {
		exitError(err)
	}
}

func exitError(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
