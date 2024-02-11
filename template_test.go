// Copyright 2013 Frederik Zipp.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"go/ast"
	"html/template"
	"testing"
)

func TestTemplateFuncs(t *testing.T) {
	data := map[string]any{
		"File":      &ast.File{},
		"FileName":  "/usr/local/go/src/pkg/fmt/format.go",
		"StdPkg":    "io/fs",
		"NonStdPkg": "golang.org/x/tools/cmd/guru",
	}
	tests := []struct {
		tmpl, want string
	}{
		{"<p>test</p>", "<p>test</p>"},
		{"{{base .FileName}}", "format.go"},
		{"{{if stdpkg .StdPkg}}ok{{else}}fail{{end}}", "ok"},
		{"{{if stdpkg .NonStdPkg}}fail{{else}}ok{{end}}", "ok"},
	}
	for _, tt := range tests {
		tmpl := template.Must(template.New("").Funcs(funcs).Parse(tt.tmpl))
		var out bytes.Buffer
		if err := tmpl.Execute(&out, data); err != nil {
			t.Errorf("Did not expect error for template %q, but got error: %q", tt.tmpl, err)
			continue
		}
		if x := out.String(); x != tt.want {
			t.Errorf("Executing template %q with data %#v resulted in %q, want %q",
				tt.tmpl, data, x, tt.want)
		}
	}
}

func TestIsStandardPackage(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"fmt", true},
		{"net/http", true},
		{"image/color/palette", true},
		{"github.com/fzipp/pythia", false},
		{"golang.org/x/tools/cmd/guru", false},
		{"main", false},
		{"foo", false},
		{"foo/bar", false},
		{"", false},
	}
	for _, tt := range tests {
		if x := isStandardPackage(tt.path); x != tt.want {
			t.Errorf("isStandardPackage(%q) = %v, want %v", tt.path, x, tt.want)
		}
	}
}
