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

type SourceCollection struct {
	Srcs, CGoSrcs, CSrcs, TSrcs []string
}

func (this *SourceCollection) Augment(other *SourceCollection) {
	this.Srcs = append(this.Srcs, other.Srcs...)
	this.CGoSrcs = append(this.CGoSrcs, other.CGoSrcs...)
	this.CSrcs = append(this.CSrcs, other.CSrcs...)
	this.TSrcs = append(this.TSrcs, other.TSrcs...)
}

func GetSourcesDepsDir(dir string) (pkg, target string, srcs *SourceCollection, deps, tdeps, funcs []string, err os.Error) {
	file, err := os.Open(dir, os.O_RDONLY, 0)
	if err != nil {
		return
	}
	finfos, err := file.Readdir(-1)
	if err != nil {
		return
	}
	deps = []string{}
	srcs = new(SourceCollection)
	for _, finfo := range finfos {
		if finfo.IsDirectory() {
			continue
		}
		name := finfo.Name
		if strings.HasSuffix(name, ".c") {
			srcs.CSrcs = append(srcs.CSrcs, name)
		}
		if strings.HasSuffix(name, ".go") {
			if name == "_testmain.go" || strings.HasSuffix(name, ".cgo1.go") || name == "_cgo_gotypes.go" {
				continue
			}
			isTest := strings.HasSuffix(name, "_test.go")
			isCgo := strings.HasPrefix(name, "cgo_")
			if isTest {
				srcs.TSrcs = append(srcs.TSrcs, name)
			} else if isCgo {
				srcs.CGoSrcs = append(srcs.CGoSrcs, name)
			} else {
				srcs.Srcs = append(srcs.Srcs, name)
			}
		}
	}

	if subsrcs, srcerr := ScanSrc(dir, "src"); srcerr == nil {
		srcs.Augment(subsrcs)
	}

	gotSources := len(srcs.Srcs) != 0 || len(srcs.TSrcs) != 0 || len(srcs.CGoSrcs) != 0

	for _, name := range srcs.Srcs {
		srcloc := path.Join(dir, name)

		var fpkg, ftarget string
		var fdeps, ffuncs []string
		fpkg, ftarget, fdeps, ffuncs, err = GetDeps(srcloc)
		if err != nil {
			return
		}
		pkg = fpkg
		if ftarget != "" {
			target = ftarget
		}
		deps = append(deps, fdeps...)
		funcs = append(funcs, ffuncs...)
	}
	for _, name := range srcs.TSrcs {
		srcloc := path.Join(dir, name)

		var fpkg, ftarget string
		var fdeps, ffuncs []string
		fpkg, ftarget, fdeps, ffuncs, err = GetDeps(srcloc)
		if err != nil {
			return
		}
		pkg = fpkg
		if ftarget != "" {
			target = ftarget
		}
		tdeps = append(tdeps, fdeps...)
		funcs = append(funcs, ffuncs...)
	}
	deps = RemoveDups(deps)
	if !gotSources {
		err = os.NewError("No source files in " + dir)
	}
	return
}

func ScanSrc(pkgdir, dir string) (srcs *SourceCollection, err os.Error) {
	file, err := os.Open(path.Join(pkgdir, dir), os.O_RDONLY, 0)
	if err != nil {
		return
	}
	names, err := file.Readdirnames(-1)
	if err != nil {
		return
	}

	srcs = new(SourceCollection)

	for _, name := range names {
		if strings.HasSuffix(name, ".c") {
			srcs.CSrcs = append(srcs.CSrcs, path.Join(dir, name))
		}
		if strings.HasSuffix(name, ".go") {
			if name == "_testmain.go" {
				continue
			}
			isTest := strings.HasSuffix(name, "_test.go")
			isCgo := strings.HasPrefix(name, "cgo_")
			if isTest {
				srcs.TSrcs = append(srcs.TSrcs, path.Join(dir, name))
			} else if isCgo {
				srcs.CGoSrcs = append(srcs.CGoSrcs, path.Join(dir, name))
			} else {
				srcs.Srcs = append(srcs.Srcs, path.Join(dir, name))
			}
		}
	}

	subdirs := GetSubDirs(path.Join(pkgdir, dir))
	for _, subdir := range subdirs {
		var subsrcs *SourceCollection
		subsrcs, err = ScanSrc(pkgdir, path.Join(dir, subdir))
		if err != nil {
			return
		}
		srcs.Augment(subsrcs)
	}
	return
}


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
