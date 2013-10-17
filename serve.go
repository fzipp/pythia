// Copyright 2013 Frederik Zipp.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/json"
	"go/ast"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"code.google.com/p/go.tools/godoc"
	"code.google.com/p/go.tools/importer"
	"code.google.com/p/go.tools/oracle"
	"github.com/fzipp/pythia/static"
)

var (
	funcs = template.FuncMap{
		"filename": func(f *ast.File) string { return imp.Fset.File(f.Pos()).Name() },
		"base":     func(path string) string { return filepath.Base(path) },
	}
	listView   = parse("list.html")
	sourceView = parse("source.html")
)

func parse(file string) *template.Template {
	return template.Must(template.New(file).Funcs(funcs).Parse(static.Files[file]))
}

func serveList(w http.ResponseWriter, req *http.Request) {
	err := listView.Execute(w, struct {
		Scope    string
		Packages []*importer.PackageInfo
	}{
		Scope:    strings.Join(args, " "),
		Packages: packages,
	})
	if err != nil {
		log.Println(err)
	}
}

func serveSource(w http.ResponseWriter, req *http.Request) {
	file := req.FormValue("file")
	if isForbidden(file) {
		errorForbidden(w)
		return
	}
	err := sourceView.Execute(w, file)
	if err != nil {
		log.Println(err)
	}
}

func serveFile(w http.ResponseWriter, req *http.Request) {
	path := req.FormValue("path")
	if isForbidden(path) {
		errorForbidden(w)
		return
	}
	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println(req.RemoteAddr, err)
		http.NotFound(w, req)
		return
	}

	var buf bytes.Buffer
	godoc.FormatText(&buf, content, -1, true, "", nil)

	buf.WriteTo(w)
}

func isForbidden(path string) bool {
	i := sort.SearchStrings(files, path)
	return i >= len(files) || files[i] != path
}

func errorForbidden(w http.ResponseWriter) {
	http.Error(w, "Forbidden", 403)
}

func serveQuery(w http.ResponseWriter, req *http.Request) {
	mode := req.FormValue("mode")
	pos := req.FormValue("pos")
	format := req.FormValue("format")
	if *verbose {
		log.Println(req.RemoteAddr, cmdLine(mode, pos, format, args))
	}
	qpos, err := oracle.ParseQueryPos(imp, pos, false)
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	mutex.Lock()
	res, err := ora.Query(mode, qpos)
	mutex.Unlock()
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	writeResult(w, res, format)
}

func writeResult(w io.Writer, res *oracle.Result, format string) {
	if format == "json" {
		b, err := json.Marshal(res)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		w.Write(b)
		return
	}
	res.WriteTo(w)
}

func serveStatic(w http.ResponseWriter, req *http.Request) {
	name := req.URL.Path
	data, ok := static.Files[name]
	if !ok {
		http.NotFound(w, req)
		return
	}
	http.ServeContent(w, req, name, time.Time{}, strings.NewReader(data))
}
