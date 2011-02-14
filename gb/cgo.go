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
	"fmt"
)
/*
CGOPKGPATH= cgo --  e1.go e2.go 
touch _cgo_run
6g -I ../_obj -o _go_.6 e3.go e1.cgo1.go e2.cgo1.go _cgo_gotypes.go
6c -FVw -I/Users/jasmuth/Documents/userland/go/pkg/darwin_amd64 _cgo_defun.c
gcc -m64 -g -fPIC -O2 -o _cgo_main.o -c   _cgo_main.c
gcc -m64 -g -fPIC -O2 -o e1.cgo2.o -c   e1.cgo2.c
gcc -m64 -g -fPIC -O2 -o e2.cgo2.o -c   e2.cgo2.c
gcc -m64 -g -fPIC -O2 -o _cgo_export.o -c   _cgo_export.c
gcc -m64 -g -fPIC -O2 -o _cgo1_.o _cgo_main.o e1.cgo2.o e2.cgo2.o _cgo_export.o  
cgo -dynimport _cgo1_.o >__cgo_import.c && mv -f __cgo_import.c _cgo_import.c
6c -FVw _cgo_import.c
rm -f _obj/e.a
gopack grc _obj/e.a _go_.6  _cgo_defun.6 _cgo_import.6 e1.cgo2.o e2.cgo2.o _cgo_export.o
mkdir -p ../_obj/; cp -f _obj/e.a ../_obj/e.a
*/

func BuildCgoPackage(pkg *Package) (err os.Error) {
	defer fmt.Println(err)
	
	if pkg.IsInGOROOT {
		return MakeBuild(pkg)
	}

	if true {
		return MakeBuild(pkg)
	}

	cgodir := path.Join(pkg.Dir, "_cgo")

	os.Mkdir(cgodir, 0755)
	
	var cgobases []string
	
	//first run cgo
	//CGOPKGPATH= cgo --  e1.go e2.go 
	cgo_argv := []string{"cgo", "--"}
	for _, cgosrc := range pkg.CGoSources {
		cgb := path.Base(cgosrc)
		cgobases = append(cgobases, cgb)
		cgd := path.Join("_cgo", cgb)
		err = Copy(pkg.Dir, cgosrc, cgd)
		cgo_argv = append(cgo_argv, cgb)
	}
	if Verbose {
		fmt.Printf("(in %s)\n", path.Join(pkg.Dir, "_cgo"))
		fmt.Printf("%v\n", cgo_argv)
	}
	
	err = RunExternal(CGoCMD, path.Join(pkg.Dir, "_cgo"), cgo_argv)
	if err != nil {
		return
	}
	
	var allsrc = []string{path.Join("_cgo", "_cgo_gotypes.go")}
	for _, src := range cgobases {
		gs := src[:len(src)-3]+".cgo1.go"
		allsrc = append(allsrc, path.Join("_cgo", gs))
	}
	allsrc = append(allsrc, pkg.GoSources...)

	pkgDest := GetRelative(pkg.Dir, GetBuildDirPkg(), CWD)
	
	
	// 6g -I ../_obj -o _go_.6 e3.go e1.cgo1.go e2.cgo1.go _cgo_gotypes.go
	err = CompilePkgSrc(pkg, allsrc, GetIBName(), pkgDest)
	if err != nil {
		return
	}
	
	// compile all the new C source
	/*
	gcc -m64 -g -fPIC -O2 -o _cgo_main.o -c   _cgo_main.c
	gcc -m64 -g -fPIC -O2 -o e1.cgo2.o -c   e1.cgo2.c
	gcc -m64 -g -fPIC -O2 -o e2.cgo2.o -c   e2.cgo2.c
	gcc -m64 -g -fPIC -O2 -o _cgo_export.o -c   _cgo_export.c
	*/
	
	gccCompile := func(src, obj string) (err os.Error) {
		gccargv := []string{"gcc", "-m64", "-g", "-fPIC", "-02", "-o", obj, "-c", src}
		if Verbose {
			fmt.Printf("%v\n", gccargv)
		}
		err = RunExternal(GCCCMD, cgodir, gccargv)
		return
	}
	
	var cobjs []string
	for _, cgb := range cgobases {
		cgc := cgb[:len(cgb)-3]+".cgo2.c"
		cgo := cgb[:len(cgb)-3]+".cgo2.o"
		cobjs = append(cobjs, cgo)
		
		err = gccCompile(cgc, cgo)
		if err != nil {
			return
		}
	}
	
	if err = gccCompile("_cgo_export.c", "_cgo_export.o"); err != nil {
		return
	}
	cobjs = append(cobjs, "_cgo_export.o")
	if err = gccCompile("_cgo_main.c", "_cgo_main.o"); err != nil {
		return
	}
	cobjs = append(cobjs, "_cgo_main.o")
	
	/* and link them
	gcc -m64 -g -fPIC -O2 -o _cgo1_.o _cgo_main.o e1.cgo2.o e2.cgo2.o _cgo_export.o  
	*/
	
	gcclargv := []string{"gcc", "-m64", "-g", "-fPIC", "-02", "-o", "_cgo1_.o"}
	gcclargv = append(gcclargv, cobjs...)
	if Verbose {
		fmt.Printf("%v\n", gcclargv)
	}
	err = RunExternal(GCCCMD, cgodir, gcclargv)
	
	//cgo -dynimport _cgo1_.o >__cgo_import.c && mv -f __cgo_import.c _cgo_import.c
	dynargv := []string{"cgo", "-dynimport", "_cgo1_.o"}
	if Verbose {
		fmt.Printf("%v > %s\n", dynargv, "__cgo_import.c")
	}
	var dump *os.File
	dump, err = os.Open(path.Join("_cgo", "__cgo_import.c"), os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return
	}
	err = RunExternalDump(CGoCMD, cgodir, dynargv, dump)
	if err != nil {
		return
	}
	
	return
}
