package main

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path"
	"strings"
	"time"
)

var fs = token.NewFileSet()

var Files = []*File{}
var Templates = []*Template{}

func parseAst(filepath string) *ast.File {
	fAst, err := parser.ParseFile(fs, filepath, nil, 0)
	deny(err)
	return fAst
}

func parse() {
	dir, err := os.Open(*fDir)
	deny(err)

	fis, err := dir.Readdir(0)
	deny(err)

	// gather template and build files
	var build []string
	var tmpl []string
	var lastModified time.Time
	var lastGenerated time.Time

fileloop:
	for _, fi := range fis {
		filepath := path.Join(*fDir, fi.Name())
		if fi.IsDir() || path.Ext(filepath) != ".go" {
			continue
		}

		f, err := parser.ParseFile(fs, filepath, nil, parser.ParseComments)
		if err != nil {
			log.Printf("failed to parse %s: %s\n", fi.Name(), err)
			continue
		}

		isTmpl := false
		isBuilding := true
		for _, group := range f.Comments {
			for _, comment := range group.List {
				cmd := strings.TrimLeft(comment.Text, "/ ")
				strings.Replace(comment.Text, " ", "", -1)
				switch cmd {
				case "+gotemplate generated":
					if lastGenerated.IsZero() || lastGenerated.After(fi.ModTime()) {
						lastGenerated = fi.ModTime()
					}
					continue fileloop

				case "+gotemplate":
					isTmpl = true

				case "+build ignore":
					isBuilding = false
				}
			}
		}
		if isTmpl {
			tmpl = append(tmpl, filepath)
		} else {
			build = append(build, filepath)
		}

		if lastModified.Before(fi.ModTime()) {
			lastModified = fi.ModTime()
		}

		if isTmpl && isBuilding {
			log.Println("adding +build ignore to", fi.Name())
			buf := bytes.NewBufferString("// +build ignore\n\n")
			f, err := os.OpenFile(filepath, os.O_RDWR, 0666)
			deny(err)
			_, err = buf.ReadFrom(f)
			deny(err)
			_, err = f.Seek(0, 0)
			deny(err)
			f.Write(buf.Bytes())
			f.Close()
		}
	}

	if lastGenerated.Before(lastModified) || len(tmpl) == 0 {
		return
	}

	for _, filepath := range tmpl {
		a := parseAst(filepath)
		file := &File{
			Path:  filepath,
			Ast:   a,
			Types: map[string]*TypeTmpl{},
			Funcs: map[string]*FuncTmpl{},
		}
		Files = append(Files, file)
		ast.Walk(file, a)
	}

	for _, filepath := range build {
		ast.Inspect(parseAst(filepath), func(node ast.Node) bool {
			id, ok := node.(*ast.Ident)
			if !ok {
				return true
			}

			parts := strings.Split(id.Name, "_")
			if len(parts) != 2 {
				return true
			}

			for _, t := range Templates {
				if !strings.HasPrefix(id.Name, t.Prefix) {
					continue
				}
				t.Implementors[parts[1]] = struct{}{}
			}

			return true
		})
	}

	finished := false
	for !finished {
		finished = true

		findPrefixes := func(t *Template) map[string]bool {
			out := map[string]bool{}
			for ref, _ := range t.References {
				parts := strings.Split(ref.Name, "_")
				if len(parts) == 2 {
					out[parts[0]] = true
				}
			}
			return out
		}

		for _, t := range Templates {
			if len(t.Implementors) == 0 {
				continue
			}

			referenced := findPrefixes(t)
			for _, t2 := range Templates {
				if referenced[t2.Prefix] {
				implloop:
					for impl, _ := range t.Implementors {
						for impl2, _ := range t2.Implementors {
							if impl == impl2 {
								continue implloop
							}
						}
						t2.Implementors[impl] = struct{}{}
						finished = false
					}
				}
			}
		}
	}
}
