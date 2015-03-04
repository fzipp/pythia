package main

import (
	"bytes"
	"go/ast"
	"testing"

	"github.com/fzipp/pythia/internal/static"
)

func TestTemplateFuncs(t *testing.T) {
	static.Files = map[string]string{
		"plain.html":   "<p>test</p>",
		"base.html":    "{{base .FileName}}",
		"stdpkg1.html": "{{if stdpkg .StdPkg}}ok{{else}}fail{{end}}",
		"stdpkg2.html": "{{if stdpkg .NonStdPkg}}fail{{else}}ok{{end}}",
	}
	data := map[string]interface{}{
		"File":      &ast.File{},
		"FileName":  "/usr/local/go/src/pkg/fmt/format.go",
		"StdPkg":    "io/ioutil",
		"NonStdPkg": "code.google.com/p/go.tools/oracle",
	}
	tests := []struct {
		file, want string
	}{
		{"plain.html", "<p>test</p>"},
		{"base.html", "format.go"},
		{"stdpkg1.html", "ok"},
		{"stdpkg2.html", "ok"},
	}
	for _, tt := range tests {
		template := parseTemplate(tt.file)
		out := new(bytes.Buffer)
		template.Execute(out, data)
		if x := out.String(); x != tt.want {
			t.Errorf("Executing template %q with data %q resulted in %q, want %q",
				tt.file, data, x, tt.want)
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
		{"code.google.com/p/go.tools/oracle", false},
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
