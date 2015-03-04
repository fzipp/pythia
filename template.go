// Copyright 2013 Frederik Zipp.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"go/ast"
	"go/build"
	"html/template"
	"path/filepath"
	"strings"

	"github.com/fzipp/pythia/internal/static"
)

var funcs = template.FuncMap{
	"filename": func(f *ast.File) string { return prog.Fset.File(f.Pos()).Name() },
	"base":     filepath.Base,
	"stdpkg":   isStandardPackage,
}

// parseTemplate reads and parses an HTML template from the static file map.
func parseTemplate(file string) *template.Template {
	return template.Must(template.New(file).Funcs(funcs).Parse(static.Files[file]))
}

// isStandardPackage returns true if the package for the given import
// path is a package of the standard library.
func isStandardPackage(path string) bool {
	p, _ := build.Import(path, "", build.FindOnly)
	return p.Goroot && p.ImportPath != "" && !strings.Contains(p.ImportPath, ".")
}
