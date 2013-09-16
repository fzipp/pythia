// Copyright 2013 Frederik Zipp.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
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

type source struct {
	FileName string
	Code     []byte
	NLines   int
}

func serveSource(w http.ResponseWriter, req *http.Request) {
	fileName := req.FormValue("file")
	format := req.FormValue("format")
	i := sort.SearchStrings(files, fileName)
	if i >= len(files) || files[i] != fileName {
		http.Error(w, "Forbidden", 403)
		return
	}
	code, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Println(req.RemoteAddr, err)
		http.NotFound(w, req)
		return
	}
	if format == "raw" {
		w.Write(code)
		return
	}
	src := source{
		FileName: fileName,
		Code:     code,
		NLines:   bytes.Count(code, []byte{'\n'}),
	}
	sourceView.Execute(w, src)
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
