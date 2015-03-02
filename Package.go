package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
)

var Packages = map[string]*Package{}

type Package struct {
	Path  string
	Name  string
	Files []*File
}

func parsePackage(dirpath string) *Package {
	if pkg, ok := Packages[dirpath]; ok {
		return pkg
	}

	fs := token.NewFileSet()
	pkgs, err := parser.ParseDir(fs, dirpath, nil, parser.ParseComments)
	if err != nil {
		if _, ok := err.(*os.PathError); !ok {
			fmt.Println(err)
		}
		return nil
	}

	switch len(pkgs) {
	case 0:
		log.Fatalf("No packages found in %s.\n", dirpath)
		return nil
	case 1:
		// continue
	default:
		var names []string
		for name, _ := range pkgs {
			names = append(names, name)
		}
		fmt.Printf("Found multiple packages in %s: %s\n",
			dirpath, strings.Join(names, ", "))
		return nil
	}

	p := &Package{
		Path: dirpath,
	}
	Packages[p.Path] = p

	for _, pkgast := range pkgs {
		p.Name = pkgast.Name
		for filepath, fast := range pkgast.Files {
			p.addFile(filepath, fast)
		}
		break
	}

	for _, f := range p.Files {
		if f.Templates != nil {
			f.parseTemplates()
		}
	}

	for _, f := range p.Files {
		if f.Templates == nil {
			p.findReferences(f)
		}
	}

	for _, f := range p.Files {
		for _, t := range f.Templates {
			if len(t.Implementors) > 0 {
				p.findReferencesRecursive(t)
			}
		}
	}
	return p
}

func (p *Package) update() {
	notChanged := true

	imports := map[*Package]struct{}{}
	for _, f := range p.Files {
		for _, pkg := range f.Imports {
			if _, ok := imports[pkg]; !ok {
				imports[pkg] = struct{}{}
				pkg.update()
			}
		}

		if f.Templates != nil {
			f.update()
			notChanged = false
		}
	}

	if notChanged {
		fmt.Printf("%s: no templates found\n", p.Path)
	}
}
