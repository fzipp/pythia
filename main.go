// Copyright 2013 Frederik Zipp.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"code.google.com/p/go.tools/importer"
	"code.google.com/p/go.tools/oracle"
	"flag"
	"fmt"
	"github.com/fzipp/pythia/static"
	"go/build"
	"go/token"
	"html/template"
	"log"
	"net/http"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
)

var (
	httpAddr   = flag.String("http", ":8080", "HTTP listen address")
	verbose    = flag.Bool("v", false, "Verbose mode: print incoming queries")
	args       []string
	files      []string
	imp        *importer.Importer
	ora        *oracle.Oracle
	mutex      sync.Mutex
	listView   = template.Must(template.New("").Parse(static.Files["list.html"]))
	sourceView = template.Must(template.New("").Parse(static.Files["source.html"]))
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
	files = scopeFiles(imp, args)

	http.HandleFunc("/", serveList)
	http.HandleFunc("/source", serveSource)
	http.HandleFunc("/file", serveFile)
	http.HandleFunc("/query", serveQuery)
	staticPrefix := "/static/"
	http.Handle(staticPrefix, http.StripPrefix(staticPrefix, http.HandlerFunc(serveStatic)))

	fmt.Printf("http://localhost%s/\n", *httpAddr)
	log.Fatal(http.ListenAndServe(*httpAddr, nil))
}

func scopeFiles(imp *importer.Importer, args []string) []string {
	files := make([]string, 0)
	imp.Fset.Iterate(func(f *token.File) bool {
		files = append(files, f.Name())
		return true
	})
	sort.Strings(files)
	return files
}

func cmdLine(mode, pos, format string) string {
	return fmt.Sprintf("oracle %s -pos=%s -format=%s %s",
		mode, pos, format, strings.Join(args, " "))
}
