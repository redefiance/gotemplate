package main

import (
	"bytes"
	"go/printer"
	"go/token"
	"os"
	"strings"
)

var buf *bytes.Buffer

func (t TypeTmpl) Print(fs *token.FileSet) bool {
	for impl, _ := range t.Implementors {
		t.Rename(impl)

		// we need to prefix type definition with "type" (not part of ast.FuncSpec)
		buf.WriteString("type ")
		printer.Fprint(buf, fs, t.Ast)
		buf.WriteString("\n\n")

		for _, method := range t.Methods {
			printer.Fprint(buf, fs, method)
			buf.WriteString("\n\n")
		}
	}
	return false
}

func (t FuncTmpl) Print(fs *token.FileSet) bool {
	for impl, _ := range t.Implementors {
		t.Rename(impl)

		printer.Fprint(buf, fs, t.Ast)
		buf.WriteString("\n\n")
	}
	return false
}

func print() {
	for _, file := range Files {
		buf = bytes.NewBufferString("// +gotemplate ignore\n\n" +
			"package " + file.Ast.Name.Name + "\n\n")

		for _, impSpec := range file.Ast.Imports {
			buf.WriteString("import ")
			printer.Fprint(buf, fs, impSpec)
			buf.WriteString("\n\n")
		}

		for _, t := range file.Funcs {
			t.Print(fs)
		}
		for _, t := range file.Types {
			t.Print(fs)
		}

		f, err := os.Create(strings.Replace(file.Path, ".go", "_impl.go", -1))
		deny(err)

		f.Write(buf.Bytes())
		f.Close()
	}
}
