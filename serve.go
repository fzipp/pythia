// Copyright 2013 Frederik Zipp.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"code.google.com/p/go.tools/oracle"
	"encoding/json"
	"github.com/fzipp/pythia/static"
	"go/build"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"
)

func serveList(w http.ResponseWriter, req *http.Request) {
	listView.Execute(w, struct {
		Scope string
		Files []string
	}{
		Scope: strings.Join(args, " "),
		Files: files,
	})
}

func serveSource(w http.ResponseWriter, req *http.Request) {
	file := req.FormValue("file")
	if isForbidden(file) {
		errorForbidden(w)
		return
	}
	sourceView.Execute(w, file)
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
	w.Write(content)
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
		log.Println(req.RemoteAddr, cmdLine(mode, pos, format))
	}
	res, err := oracle.Query(args, mode, pos, nil, &build.Default)
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
