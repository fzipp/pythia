// Copyright 2013 Frederik Zipp.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strings"

	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/godoc"
)

//go:embed static/*
var staticFiles embed.FS

// serveIndex delivers the scope index page, which is the first
// page presented to the user.
func serveIndex(w http.ResponseWriter, req *http.Request) {
	err := templates.Lookup("index.html").Execute(w, struct {
		Scope    string
		Packages []*loader.PackageInfo
	}{
		Scope:    strings.Join(args, " "),
		Packages: packages,
	})
	if err != nil {
		log.Println(err)
	}
}

// serveSource delivers the source view page, which is the main
// workspace of the tool, where the user creates the queries to
// the guru and browses their results.
//
// The request parameter 'file' determines the source file to be
// shown initially, e.g. "/path/to/file.go". The contents of the
// file are not loaded in this request, but in a subsequent
// asynchronous request handled by serveFile.
//
// Returns a "403 Forbidden" status code if the requested file
// is not within the import scope.
func serveSource(w http.ResponseWriter, req *http.Request) {
	file := req.FormValue("file")
	if isForbidden(file) {
		errorForbidden(w)
		return
	}
	err := templates.Lookup("source.html").Execute(w, file)
	if err != nil {
		log.Println(err)
	}
}

// serveFile delivers an HTML fragment of a Go source file with
// highlighted comments and an (optional) highlighted selection.
// The request parameters are:
//
//   path: "/path/to/file.go"
//   s: optional selection range like "line.col-line.col", e.g. "24.4-25.10"
//
// Returns a "403 Forbidden" status code if the requested file
// is not within the import scope, or a "404 Not Found" if the
// file can't be read.
func serveFile(w http.ResponseWriter, req *http.Request) {
	path := req.FormValue("path")
	if isForbidden(path) {
		errorForbidden(w)
		return
	}
	content, err := os.ReadFile(path)
	if err != nil {
		log.Println(req.RemoteAddr, err)
		http.NotFound(w, req)
		return
	}

	var sel godoc.Selection
	s, err := parseSelection(req.FormValue("s"))
	if err == nil {
		offsets := s.byteOffsetsIn(content)
		sel = godoc.RangeSelection(offsets)
	}

	var buf bytes.Buffer
	godoc.FormatText(&buf, content, -1, true, "", sel)

	buf.WriteTo(w)
}

// isForbidden checks if the given file path is in the file set of the
// imported scope and returns true if not, otherwise false.
func isForbidden(path string) bool {
	// files must be sorted!
	i := sort.SearchStrings(files, path)
	return i >= len(files) || files[i] != path
}

func errorForbidden(w http.ResponseWriter) {
	http.Error(w, "Forbidden", http.StatusForbidden)
}

// serveQuery executes a query to the guru and delivers the results
// in the specified format. The request parameters are:
//
//   mode: e.g. "describe", "callers", "freevars", ...
//   pos: file name with byte offset(s), e.g. "/path/to/file.go:#1457,#1462"
//   format: "json" or "plain", no "xml" at the moment
//
// If the application was launched in verbose mode, each query will be
// logged like an invocation of the guru command.
func serveQuery(w http.ResponseWriter, req *http.Request) {
	mode := req.FormValue("mode")
	pos := req.FormValue("pos")
	format := req.FormValue("format")
	if format != "json" && format != "plain" {
		fmt.Println("Warning: incorrect format:", mode, pos, format)
	}
	// Call guru
	var args []string
	if format == "json" {
		args = append(args, "-json")
	}
	args = append(args, "-scope", scope, mode, pos)
	if *verbose {
		log.Println("guru", strings.Join(args, " "))
	}
	out, err := exec.Command(guruPath, args...).Output()
	if err != nil {
		log.Println("launch guru command:", err)
		http.Error(w, "guru command failed", http.StatusInternalServerError)
		return
	}
	w.Write(out)
}
