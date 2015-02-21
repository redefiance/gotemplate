package main

import (
	"fmt"
	"go/ast"
	"go/printer"
	"os"
	"path"
	"strings"
)

type File struct {
	Path      string
	Ast       *ast.File
	Imports   map[string]*Package
	Templates map[string]*Template
}

func (p *Package) addFile(filepath string, fast *ast.File) {
	f := &File{
		Path:    filepath,
		Ast:     fast,
		Imports: map[string]*Package{},
	}

	ignoreBuild := false
	for _, group := range f.Ast.Comments {
		for _, comment := range group.List {
			switch strings.TrimLeft(comment.Text, " /") {
			case "+build ignore":
				ignoreBuild = true
			case "+gotemplate ignore":
				ignoreBuild = true
			case "+gotemplate":
				f.Templates = map[string]*Template{}
			}
		}
	}

	if ignoreBuild && f.Templates == nil {
		return
	}
	if !ignoreBuild && f.Templates != nil {
		// TODO
	}

	for _, imp := range f.Ast.Imports {
		pkgpath := path.Join(GOPATH, "src", strings.Trim(imp.Path.Value, "\""))
		if pkg := parsePackage(pkgpath); pkg != nil {
			var name string
			if imp.Name != nil {
				name = imp.Name.Name
			} else {
				name = path.Base(pkg.Path)
			}
			f.Imports[name] = pkg
		}
	}

	p.Files = append(p.Files, f)
}

func (f *File) update() {
	if f.Templates == nil {
		return
	}

	filepath := strings.Replace(f.Path, ".go", "_impl.go", 1)
	// TODO sanity check

	fmt.Println("update", filepath)
	buf, err := os.Create(filepath)
	deny(err)

	buf.WriteString("// +gotemplate ignore\n\n")
	buf.WriteString("package " + f.Ast.Name.Name + "\n\n")

	if len(f.Ast.Imports) > 0 {
		buf.WriteString("import (\n")
		for _, imp := range f.Ast.Imports {
			buf.WriteString("  ")
			printer.Fprint(buf, FS, imp)
			buf.WriteString("\n")
		}
		buf.WriteString(")\n\n")
	}
	// buf := bytes.NewBufferString("// +gotemplate ignore\n\n")
	for _, t := range f.Templates {
		for impl, _ := range t.Implementors {
			for ref, _ := range t.References {
				parts := strings.Split(ref.Name, "_")
				parts[len(parts)-1] = impl
				ref.Name = strings.Join(parts, "_")
			}
			if _, ok := t.Ast.(*ast.TypeSpec); ok {
				buf.WriteString("type ")
			}
			printer.Fprint(buf, FS, t.Ast)
			buf.WriteString("\n\n")

			for _, method := range t.Methods {
				printer.Fprint(buf, FS, method)
				buf.WriteString("\n\n")
			}
		}
	}

	// fmt.Println(buf.String())

	buf.Close()
}
