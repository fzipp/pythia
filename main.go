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
	"go/ast"
	"go/build"
	"go/token"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
)

var (
	httpAddr   = flag.String("http", ":8080", "HTTP listen address")
	verbose    = flag.Bool("v", false, "Verbose mode: print incoming queries")
	open       = flag.Bool("open", true, "Try to open browser")
	args       []string
	files      []string
	packages   []*importer.PackageInfo
	imp        *importer.Importer
	ora        *oracle.Oracle
	mutex      sync.Mutex
	listView   *template.Template
	sourceView *template.Template

	funcMap = template.FuncMap{
		"filename": func(f *ast.File) string { return imp.Fset.File(f.Pos()).Name() },
		"base":     func(path string) string { return filepath.Base(path) },
	}
)

func init() {
	if os.Getenv("GOMAXPROCS") == "" {
		n := runtime.NumCPU()
		if n < 4 {
			n = 4
		}
		runtime.GOMAXPROCS(n)
	}
	listView = template.New("").Funcs(funcMap)
	template.Must(listView.Parse(static.Files["list.html"]))
	sourceView = template.Must(template.New("").Parse(static.Files["source.html"]))
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
	sort.Sort(PackageInfos(packages))

	http.HandleFunc("/", serveList)
	http.HandleFunc("/source", serveSource)
	http.HandleFunc("/file", serveFile)
	http.HandleFunc("/query", serveQuery)
	staticPrefix := "/static/"
	http.Handle(staticPrefix, http.StripPrefix(staticPrefix, http.HandlerFunc(serveStatic)))

	if *open {
		url := fmt.Sprintf("http://localhost%s/", *httpAddr)
		if !startBrowser(url) {
			fmt.Println(url)
		}
	}
	log.Fatal(http.ListenAndServe(*httpAddr, nil))
}

type PackageInfos []*importer.PackageInfo

func (p PackageInfos) Len() int           { return len(p) }
func (p PackageInfos) Less(i, j int) bool { return p[i].Pkg.Path() < p[j].Pkg.Path() }
func (p PackageInfos) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

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

func cmdLine(mode, pos, format string) string {
	return fmt.Sprintf("oracle %s -pos=%s -format=%s %s",
		mode, pos, format, strings.Join(args, " "))
}
