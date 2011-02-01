/* 
   Copyright 2011 John Asmuth

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package main

import (
	"os"
	"strings"
	"go/parser"
	"go/token"
	"go/ast"
)

func GetDeps(source string) (pkg, target string, deps, funcs []string, err os.Error) {
	var file *ast.File
	file, err = parser.ParseFile(token.NewFileSet(), source, nil, parser.ParseComments)
	if err != nil {
		println(err.String())
		BrokenPackages++
		return
	}

	w := &Walker{"", "", 0, []string{}, []string{}, strings.HasSuffix(source, "_test.go")}

	ast.Walk(w, file)

	deps = w.Deps
	pkg = w.Name
	target = w.Target
	funcs = w.Funcs

	return
}

func RemoveDups(list []string) (newlist []string) {
	m := make(map[string]bool)
	for _, item := range list {
		m[item] = true
	}
	newlist = make([]string, 0)
	for item, _ := range m {
		newlist = append(newlist, item)
	}
	return
}

type Walker struct {
	Name   string
	Target string
	pkgPos token.Pos
	Deps   []string
	Funcs  []string
	ScanFuncs bool
}

func (w *Walker) Visit(node ast.Node) (v ast.Visitor) {
	switch n := node.(type) {
	case *ast.File:
		w.Name = n.Name.Name
		w.pkgPos = n.Package
		return w
	case *ast.ImportSpec:
		w.Deps = append(w.Deps, string(n.Path.Value))
		return nil
	case *ast.CommentGroup:
		return w
	case *ast.Comment:
		if n.Pos() < w.pkgPos {
			text := string(n.Text)
			if strings.HasPrefix(text, "//target:") {
				w.Target = text[len("//target:"):len(text)]
			}
		}
		return nil
	case *ast.GenDecl:
		return w
	case *ast.FuncDecl:
		if w.ScanFuncs {
			fdecl, ok := node.(*ast.FuncDecl)
			if ok {
				w.Funcs = append(w.Funcs, fdecl.Name.Name)
			}
		}
		return nil
	}
	return nil
}
