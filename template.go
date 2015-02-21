package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

type Template struct {
	Prefix       string
	Param        string
	Implementors map[string]struct{}
	References   map[*ast.Ident]struct{}
}

type TypeTmpl struct {
	*Template
	Ast     *ast.TypeSpec
	Methods []*ast.FuncDecl
}

type FuncTmpl struct {
	*Template
	Ast *ast.FuncDecl
}

type File struct {
	Path  string
	Ast   *ast.File
	Types map[string]*TypeTmpl
	Funcs map[string]*FuncTmpl
}

func (t *Template) Rename(impl string) {
	fmt.Println("implement " + t.Prefix + "_" + impl)
	for ref, _ := range t.References {
		parts := strings.Split(ref.Name, "_")
		if len(parts) == 1 {
			ref.Name = impl
		} else {
			parts[1] = impl
			ref.Name = strings.Join(parts, "_")
		}
	}
}

func (t *Template) parseScope(node ast.Node) {
	ast.Inspect(node, func(node ast.Node) bool {
		id, ok := node.(*ast.Ident)
		if !ok {
			return true
		}
		if id.Name == t.Param || strings.HasSuffix(id.Name, "_"+t.Param) {
			t.References[id] = struct{}{}
		}
		return true
	})
}

// extracts prefix and parameter list from an identifier
func newTemplate(id *ast.Ident) (tmpl *Template) {
	parts := strings.Split(id.Name, "_")
	if len(parts) != 2 {
		fmt.Printf("%s: only one template parameter allowed\n", id.Name)
		return nil
	}

	tmpl = &Template{
		Prefix:       parts[0],
		Param:        parts[1],
		Implementors: map[string]struct{}{},
		References:   map[*ast.Ident]struct{}{},
	}
	Templates = append(Templates, tmpl)
	return
}

// scans top-level-definitions of a file for type and func definitions
func (v *File) Visit(node ast.Node) (next ast.Visitor) {
	getTypeDef := func(id *ast.Ident) *TypeTmpl {
		name := id.Name
		if _, ok := v.Types[name]; !ok {
			v.Types[name] = &TypeTmpl{Template: newTemplate(id)}
		}
		return v.Types[name]
	}

	switch e := node.(type) {
	case *ast.File:
		// entry point, scan the file
		return v

	case *ast.GenDecl:
		// scan all type declarations
		if e.Tok == token.TYPE {
			return v
		}

	case *ast.TypeSpec:
		t := getTypeDef(e.Name)
		t.Ast = e
		t.References[e.Name] = struct{}{}
		t.parseScope(e)

	case *ast.FuncDecl:
		if e.Recv == nil { // function
			t := &FuncTmpl{
				Template: newTemplate(e.Name),
				Ast:      e,
			}
			t.parseScope(e)
			v.Funcs[e.Name.Name] = t
		} else { // method
			ast.Inspect(e.Recv.List[0].Type, func(node ast.Node) bool {
				if id, ok := node.(*ast.Ident); ok {
					t := getTypeDef(id)
					t.Methods = append(t.Methods, e)
					t.References[id] = struct{}{}
					t.parseScope(e)
				}
				return true
			})
		}
	}
	return nil
}
