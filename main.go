// Copyright 2013 Frederik Zipp.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"code.google.com/p/go.tools/importer"
	"code.google.com/p/go.tools/oracle"
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
	"sort"
	"strings"
	"sync"
)

var (
	httpAddr = flag.String("http", ":8080", "HTTP listen address")
	verbose  = flag.Bool("v", false, "Verbose mode: print incoming queries")
	open     = flag.Bool("open", true, "Try to open browser")
	args     []string
	files    []string
	packages []*importer.PackageInfo
	imp      *importer.Importer
	ora      *oracle.Oracle
	mutex    sync.Mutex
)

func init() {
	if os.Getenv("GOMAXPROCS") == "" {
		n := runtime.NumCPU()
		if n < 4 {
			n = 4
		}
		runtime.GOMAXPROCS(n)
	}
}

const usage = `Web frontend for Go source code oracle.
Usage: pythia [<flag> ...] <args> ...
Use -help flag to display options.
` + importer.InitialPackagesUsage

func main() {
	flag.Parse()
	args = flag.Args()
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}

	var err error
	imp = importer.New(&importer.Config{Build: &build.Default})
	ora, err = oracle.New(imp, args, nil, false)
	if err != nil {
		log.Fatal(err)
	}
	files = scopeFiles(imp)
	packages = imp.AllPackages()
	sort.Sort(byPath(packages))

	registerHandlers()

	srv := &http.Server{Addr: *httpAddr}
	l, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		log.Fatal(err)
	}
	if *open {
		url := fmt.Sprintf("http://localhost%s/", *httpAddr)
		if !startBrowser(url) {
			fmt.Println(url)
		}
	}
	log.Fatal(srv.Serve(l))
}

func registerHandlers() {
	http.HandleFunc("/", serveList)
	http.HandleFunc("/source", serveSource)
	http.HandleFunc("/file", serveFile)
	http.HandleFunc("/query", serveQuery)
	staticPrefix := "/static/"
	http.Handle(staticPrefix, http.StripPrefix(staticPrefix, http.HandlerFunc(serveStatic)))
}

// byPath makes a slice of package infos sortable by package path.
type byPath []*importer.PackageInfo

func (p byPath) Len() int           { return len(p) }
func (p byPath) Less(i, j int) bool { return p[i].Pkg.Path() < p[j].Pkg.Path() }
func (p byPath) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// scopeFiles returns a new slice containing the full paths of all the files
// imported by an importer, sorted in increasing order.
func scopeFiles(imp *importer.Importer) []string {
	files := make([]string, 0)
	imp.Fset.Iterate(func(f *token.File) bool {
		files = append(files, f.Name())
		return true
	})
	sort.Strings(files)
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

// cmdLine returns what the command line would look like if the oracle was
// invoked via command line with the given arguments.
func cmdLine(mode, pos, format string, scope []string) string {
	return fmt.Sprintf("oracle %s -pos=%s -format=%s %s",
		mode, pos, format, strings.Join(scope, " "))
}
