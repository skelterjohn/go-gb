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
	"path"
	"strings"
	"go/parser"
	"go/token"
	"go/ast"
)

func PkgExistsInGOROOT(target string) bool {
	if target[0] == '"' {
		target = target[1:len(target)]
	}
	if target[len(target)-1] == '"' {
		target = target[0 : len(target)-1]
	}

	pkgbin := path.Join(GetInstallDirPkg(), target)
	pkgbin += ".a"
	
	_, err := os.Stat(pkgbin)
	
	return err == nil
}

func FilterFlag(src string) bool {

	os_flags := []string{"windows", "darwin", "freebsd", "bsd", "linux"}
	arch_flags := []string{"amd64", "386", "arm"}
	for _, flag := range os_flags {
		if strings.Contains(src, "_"+flag) && GOOS != flag {
			return false
		}
	}
	for _, flag := range arch_flags {
		if strings.Contains(src, "_"+flag) && GOARCH != flag {
			return false
		}
	}
	if strings.Contains(src, "_unix") && 
		!(GOOS == "darwin" || GOOS == "freebsd" || GOOS == "bsd" || GOOS == "linux") {
		return false
	}
	
	return true
}

type SourceWalker struct {
	root string
	srcroot string
	srcs []string
	tsrcs []string
	csrcs []string
	cgosrcs []string
}
func (this *SourceWalker) VisitDir(dpath string, f *os.FileInfo) bool {
	return dpath == this.root || strings.HasPrefix(dpath, this.srcroot)
}
func (this *SourceWalker) VisitFile(fpath string, f *os.FileInfo) {
	if !FilterFlag(fpath) {
		return
	}
	if strings.HasSuffix(fpath, "_testmain.go") {
		return
	}
	rootl := len(this.root)+1
	if this.root != "." {
		fpath = fpath[rootl:len(fpath)]
	}
	if strings.HasSuffix(fpath, ".go") {
		if strings.HasSuffix(fpath, "_test.go") {
			this.tsrcs = append(this.tsrcs, fpath)
		} else if strings.HasPrefix(fpath, "cgo_") {
			this.cgosrcs = append(this.cgosrcs, fpath)
		} else {
			this.srcs = append(this.srcs, fpath)
		}
	}
	if strings.HasSuffix(fpath, ".c") {
		this.csrcs = append(this.csrcs, fpath)
	}
}

func GetDepsMany(dir string, srcs []string) (err os.Error) {
	fset := token.NewFileSet()
	filenames := make([]string, len(srcs))
	for i, src := range srcs {
		filenames[i] = path.Join(dir, src)
	}
	pkgs, err := parser.ParseFiles(fset, filenames, parser.ParseComments)
	for _, pkg := range pkgs {
		w := &Walker{"", "", 0, []string{}, []string{}, false}

		ast.Walk(w, pkg)

	}
	return
}

func GetDeps(source string) (pkg, target string, deps, funcs []string, err os.Error) {
	isTest := strings.HasSuffix(source, "_test.go") && Test
	var file *ast.File
	flag := parser.ParseComments
	if !isTest {
		flag = flag | parser.ImportsOnly
	}
	file, err = parser.ParseFile(token.NewFileSet(), source, nil, flag)
	if err != nil {
		println(err.String())
		BrokenPackages++
		return
	}

	w := &Walker{"", "", 0, []string{}, []string{}, isTest}

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
	case *ast.Comment:
		if n.Pos() < w.pkgPos {
			text := string(n.Text)
			if strings.HasPrefix(text, "//target:") {
				w.Target = text[len("//target:"):len(text)]
			}
		}
		return nil
	case *ast.FuncDecl:
		if w.ScanFuncs {
			fdecl, ok := node.(*ast.FuncDecl)
			if ok {
				w.Funcs = append(w.Funcs, fdecl.Name.Name)
			}
		}
		return nil
	case *ast.GenDecl, *ast.CommentGroup:
		return w
	}
	return nil
}
