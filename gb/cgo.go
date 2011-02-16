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

var TestCGO = true

func BuildCgoPackage(pkg *Package) (err os.Error) {
	//defer fmt.Println(err)

	if pkg.IsInGOROOT {
		return MakeBuild(pkg)
	}

	if !TestCGO {
		return MakeBuild(pkg)
	}

	cgodir := path.Join(pkg.Dir, "_cgo")

	if Verbose {
		fmt.Printf("Creating directory %s\n", cgodir)
	}
	err = os.MkdirAll(cgodir, 0755)
	if err != nil {
		return
	}

	var cgobases []string

	//first run cgo
	//CGOPKGPATH= cgo --  e1.go e2.go 
	cgo_argv := []string{"cgo", "--", "-I.."}
	for _, cgosrc := range pkg.CGoSources {
		cgb := path.Base(cgosrc)
		cgobases = append(cgobases, cgb)
		cgd := path.Join("_cgo", cgb)
		err = Copy(pkg.Dir, cgosrc, cgd)
		cgo_argv = append(cgo_argv, cgb)
	}
	if Verbose {
		fmt.Printf("%s:", cgodir)
		fmt.Printf("%v\n", cgo_argv)
	}
	err = RunExternal(CGoCMD, cgodir, cgo_argv)
	if err != nil {
		return
	}

	var allsrc = []string{path.Join("_cgo", "_cgo_gotypes.go")}
	for _, src := range cgobases {
		gs := src[:len(src)-3] + ".cgo1.go"
		allsrc = append(allsrc, path.Join("_cgo", gs))
	}
	allsrc = append(allsrc, pkg.GoSources...)

	pkgDest := GetRelative(pkg.Dir, GetBuildDirPkg(), CWD)

	// 6g -I ../_obj -o _go_.6 e3.go e1.cgo1.go e2.cgo1.go _cgo_gotypes.go
	err = CompilePkgSrc(pkg, allsrc, GetIBName(), pkgDest)
	if err != nil {
		return
	}

	//6c -FVw -I/Users/jasmuth/Documents/userland/go/pkg/darwin_amd64 _cgo_defun.c
	cdefargv := []string{GetCCompilerName(), "-FVw", "-I" + GetInstallDirPkg(), "_cgo_defun.c"}
	if Verbose {
		fmt.Printf("%s:", cgodir)
		fmt.Printf("%v\n", cdefargv)
	}
	err = RunExternal(CCMD, cgodir, cdefargv)
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
		gccargv := []string{"gcc", "-I.."}
		gccargv = append(gccargv, []string{"-m64", "-g", "-fPIC", "-O2", "-o", obj, "-c"}...)
		gccargv = append(gccargv, pkg.CGoCFlags[pkg.Name]...)
		gccargv = append(gccargv, src)
		if Verbose {
			fmt.Printf("%s:", cgodir)
			fmt.Printf("%v\n", gccargv)
		}
		err = RunExternal(GCCCMD, cgodir, gccargv)
		return
	}
	var cobjs []string
	for _, cgb := range cgobases {
		cgc := cgb[:len(cgb)-3] + ".cgo2.c"
		cgo := cgb[:len(cgb)-3] + ".cgo2.o"
		cobjs = append(cobjs, cgo)

		err = gccCompile(cgc, cgo)
		if err != nil {
			return
		}
	}
	
	for _, csrc := range pkg.CSrcs {
		cobj := csrc[:len(csrc)-2]+".o"
		cobj = path.Base(cobj)
		cobjs = append(cobjs, cobj)
		relsrc := GetRelative("_cgo", csrc, path.Join(CWD, pkg.Dir))
		err = gccCompile(relsrc, cobj)
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

	/* and link them
	gcc -m64 -g -fPIC -O2 -o _cgo1_.o _cgo_main.o e1.cgo2.o e2.cgo2.o _cgo_export.o  
	*/
	gcclargv := []string{"gcc"}
	gcclargv = append(gcclargv, []string{"-m64", "-g", "-fPIC", "-O2", "-o", "_cgo1_.o"}...)
	gcclargv = append(gcclargv, "_cgo_main.o")
	gcclargv = append(gcclargv, cobjs...)
	gcclargv = append(gcclargv, pkg.CGoLDFlags[pkg.Name]...)
	if Verbose {
		fmt.Printf("%s:", cgodir)
		fmt.Printf("%v\n", gcclargv)
	}
	err = RunExternal(GCCCMD, cgodir, gcclargv)
	if err != nil {
		return
	}
	
	//cgo -dynimport _cgo1_.o >__cgo_import.c && mv -f __cgo_import.c _cgo_import.c
	dynargv := []string{"cgo", "-dynimport", "_cgo1_.o"}
	if Verbose {
		fmt.Printf("%s:", cgodir)
		fmt.Printf("%v > %s\n", dynargv, "__cgo_import.c")
	}

	var dump *os.File
	dump, err = os.Open(path.Join(cgodir, "__cgo_import.c"), os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return
	}
	err = RunExternalDump(CGoCMD, cgodir, dynargv, dump)
	if err != nil {
		return
	}
	dump.Close()

	//mv __cgo_import.c _cgo_import.c
	if Verbose {
		fmt.Printf("%s:", cgodir)
		fmt.Printf("Moving __cgo_import.c to _cgo_import.c\n")
	}
	err = os.Rename(path.Join(cgodir, "__cgo_import.c"), path.Join(cgodir, "_cgo_import.c"))
	if err != nil {
		return
	}

	/* compile the C bits
	6c -FVw _cgo_import.c
	*/
	ccargv := []string{GetCCompilerName(), "-FVw", "_cgo_import.c"}
	if Verbose {
		fmt.Printf("%s:", cgodir)
		fmt.Printf("%v\n", ccargv)
	}
	err = RunExternal(CCMD, cgodir, ccargv)
	if err != nil {
		return
	}

	/*clean/link
	rm -f _obj/e.a
	gopack grc _obj/e.a _go_.6  _cgo_defun.6 _cgo_import.6 e1.cgo2.o e2.cgo2.o _cgo_export.o
	*/
	dst := GetRelative(".", pkg.ResultPath, CWD)
	reldst := GetRelative(pkg.Dir, pkg.ResultPath, CWD)
	dstDir, _ := path.Split(dst)
	if Verbose {
		fmt.Printf("Creating directory %s\n", dstDir)
	}
	err = os.MkdirAll(dstDir, 0755)
	if err != nil {
		return
	}
	if Verbose {
		fmt.Printf("Removing %s\n", dst)
	}
	os.Remove(dst)

	relobjs := []string{}
	for _, cobj := range cobjs {
		relobjs = append(relobjs, path.Join("_cgo", cobj))
	}
	packargv := []string{"gopack", "grc", reldst, GetIBName(),
		path.Join("_cgo", "_cgo_defun"+GetObjSuffix()),
		path.Join("_cgo", "_cgo_import"+GetObjSuffix())}
	packargv = append(packargv, relobjs...)
	if Verbose {
		fmt.Printf("%v\n", packargv)
	}
	err = RunExternal(PackCMD, pkg.Dir, packargv)
	return
}

func CleanCGoPackage(pkg *Package) (err os.Error) {
	if !TestCGO {
		err = MakeClean(pkg)
		return
	}

	if Verbose {
		fmt.Printf(" Removing %s\n", path.Join(pkg.Dir, "_cgo"))
	}
	os.RemoveAll(path.Join(pkg.Dir, "_cgo"))

	return
}
