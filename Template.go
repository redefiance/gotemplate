package main

import (
	"go/ast"
	"strings"
)

type Template struct {
	Prefix       string
	Param        string
	Implementors map[string]struct{}
	Ast          ast.Node
	References   map[*ast.Ident]struct{}
	Methods      []*ast.FuncDecl
}

func (f *File) parseName(id *ast.Ident) *Template {
	var (
		prefix = id.Name
		param  = ""
		parts  = strings.Split(id.Name, "_")
	)
	if len(parts) == 2 {
		prefix = parts[0]
		param = parts[1]
	}

	if param == "" {
		return nil
	}
	t, ok := f.Templates[prefix]
	if !ok {
		t = &Template{
			Prefix:       prefix,
			Param:        param,
			Implementors: map[string]struct{}{},
			References:   map[*ast.Ident]struct{}{},
		}
		f.Templates[prefix] = t
	}
	return t
}

func (t *Template) parseBody(node ast.Node) {
	ast.Inspect(node, func(node ast.Node) bool {
		switch e := node.(type) {
		case *ast.Ident:
			if e.Name == t.Prefix+"_"+t.Param {
				t.References[e] = struct{}{}
			} else if e.Name == t.Param || strings.HasSuffix(e.Name, "_"+t.Param) {
				t.References[e] = struct{}{}
			}
		}
		return true
	})
}

func (f *File) parseFunc(decl *ast.FuncDecl) {
	var t *Template
	if decl.Recv == nil { // function
		t = f.parseName(decl.Name)
		t.Ast = decl
	} else { // method
		ast.Inspect(decl.Recv.List[0].Type, func(node ast.Node) bool {
			switch e := node.(type) {
			case *ast.Ident:
				t = f.parseName(e)
				return false
			default:
				return true
			}
		})
	}
	if t == nil {
		return
	}
	if decl.Recv != nil {
		t.Methods = append(t.Methods, decl)
	}
	t.parseBody(decl)
}

func (f *File) parseType(decl *ast.TypeSpec) {
	if t := f.parseName(decl.Name); t != nil {
		t.Ast = decl
		t.parseBody(decl)
	}
}

func (f *File) parseTemplates() {
	ast.Inspect(f.Ast, func(node ast.Node) bool {
		switch e := node.(type) {
		case *ast.File:
			return true
		case *ast.GenDecl:
			return true
		case *ast.FuncDecl:
			f.parseFunc(e)
		case *ast.TypeSpec:
			f.parseType(e)
		}
		return false
	})
}
