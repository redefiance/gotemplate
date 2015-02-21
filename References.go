package main

import (
	"go/ast"
	"strings"
)

func (p *Package) findReferences(f *File) {
	check := func(id *ast.Ident) {
		parts := strings.Split(id.Name, "_")
		if len(parts) == 2 {
			for _, f := range p.Files {
				if f.Templates == nil {
					continue
				}
				if t, ok := f.Templates[parts[0]]; ok {
					t.Implementors[parts[1]] = struct{}{}
				}
			}
		}
	}

	ast.Inspect(f.Ast, func(node ast.Node) bool {
		switch e := node.(type) {
		case *ast.Ident:
			check(e)
		}
		return true
	})
}

func (p *Package) findReferencesRecursive(t *Template) {
	for ref, _ := range t.References {
		for _, f := range p.Files {
			for _, t2 := range f.Templates {
				if ref.Name == t2.Prefix+"_"+t.Param {
					updated := false
					for impl, _ := range t.Implementors {
						if _, ok := t2.Implementors[impl]; !ok {
							t2.Implementors[impl] = struct{}{}
							updated = true
						}
					}
					if updated {
						p.findReferencesRecursive(t2)
					}
					break
				}
			}
		}
	}
}
