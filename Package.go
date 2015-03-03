package main

import (
	"fmt"
	"go/parser"
	"go/token"
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

	for pkgname, pkgast := range pkgs {
		if strings.HasSuffix(pkgname, "_test") {
			continue
		}

		p := &Package{
			Path: dirpath,
		}
		Packages[p.Path] = p

		p.Name = pkgast.Name
		for filepath, fast := range pkgast.Files {
			p.addFile(filepath, fast)
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

	// no package found?
	return nil
}

func (p *Package) generate() {
	numImplementations := 0

	imports := map[*Package]struct{}{}
	for _, f := range p.Files {
		for _, pkg := range f.Imports {
			if _, ok := imports[pkg]; !ok {
				imports[pkg] = struct{}{}
				pkg.generate()
			}
		}

		if f.Templates != nil {
			numImplementations += f.generate()
		}
	}

	fmt.Printf("%s: %d template instantiations generated\n", p.Path, numImplementations)
}
