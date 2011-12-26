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
	"fmt"
	"os"
	"path/filepath"
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
/*
_CGO_CFLAGS_386=-m32
_CGO_CFLAGS_amd64=-m64
_CGO_LDFLAGS_freebsd=-shared -lpthread -lm
_CGO_LDFLAGS_linux=-shared -lpthread -lm
_CGO_LDFLAGS_darwin=-dynamiclib -Wl,-undefined,dynamic_lookup
_CGO_LDFLAGS_windows=-shared -lm -mthreads
*/
var TestCGO = true

func BuildCgoPackage(pkg *Package) (err error) {
	//defer fmt.Println(err)

	/*
		if pkg.IsInGOROOT {
			return MakeBuild(pkg)
		}
	*/

	if !TestCGO {
		return MakeBuild(pkg)
	}

	var CFLAGS []string
	var LDFLAGS []string

	switch GOARCH {
	case "amd64":
		CFLAGS = []string{"-m64"}
	default:
		CFLAGS = []string{"-m32"}
	}

	switch GOOS {
	case "freebsd":
	case "linux":
		LDFLAGS = []string{"-shared", "-lpthread", "-lm"}
	case "darwin":
		LDFLAGS = []string{"-dynamiclib", "-Wl,-undefined,dynamic_lookup"}
	case "windows":
		LDFLAGS = []string{"-shared", "-lm", "-mthreads"}
	}

	_ = LDFLAGS // apparently the makefile doesn't use them...

	cgodir := filepath.Join(pkg.Dir, "_cgo")

	if Verbose {
		fmt.Printf("Creating directory %s\n", cgodir)
	}
	err = os.MkdirAll(cgodir, 0755)
	if err != nil {
		return
	}

	defer func() {
		if Verbose {
			fmt.Printf("Removing directory %s\n", cgodir)
		}
		os.RemoveAll(cgodir)
	}()

	var cgobases []string

	//first run cgo
	//CGOPKGPATH= cgo --  e1.go e2.go
	cgo_argv := []string{"cgo", "--", "-I.."}
	for _, cgosrc := range pkg.CGoSources {
		cgb := filepath.Base(cgosrc)
		cgobases = append(cgobases, cgb)
		cgd := filepath.Join("_cgo", cgb)
		err = Copy(pkg.Dir, cgosrc, cgd)
		cgo_argv = append(cgo_argv, cgb)
	}
	if len(pkg.CGoSources) != 0 {
		if Verbose {
			fmt.Printf("%s:", cgodir)
		}
		err = RunExternal(CGoCMD, cgodir, cgo_argv)
		if err != nil {
			return
		}
	}

	var allsrc []string
	if len(pkg.CGoSources) != 0 {
		allsrc = append(allsrc, filepath.Join("_cgo", "_obj", "_cgo_gotypes.go"))
	}
	for _, src := range cgobases {
		gs := src[:len(src)-3] + ".cgo1.go"
		allsrc = append(allsrc, filepath.Join("_cgo", "_obj", gs))
	}
	allsrc = append(allsrc, pkg.PkgSrc[pkg.Name]...)

	pkgDest := GetRelative(pkg.Dir, GetBuildDirPkg(), CWD)

	var testDest string
	if pkg.InTestData != "" {
		tdBuildDir := filepath.Join(pkg.InTestData, GetBuildDirPkg())
		testDest = GetRelative(pkg.Dir, tdBuildDir, CWD)
	}

	ibname := GetIBName()

	// 6g -I ../_obj -o _go_.6 e3.go e1.cgo1.go e2.cgo1.go _cgo_gotypes.go
	err = CompilePkgSrc(pkg, allsrc, ibname, pkgDest, testDest)
	if err != nil {
		return
	}

	defer func() {
		if Verbose {
			fmt.Printf("Removing %s\n", filepath.Join(pkg.Dir, ibname))
		}
		os.Remove(filepath.Join(pkg.Dir, ibname))
	}()

	//6c -FVw -I/Users/jasmuth/Documents/userland/go/pkg/darwin_amd64 _cgo_defun.c

	gorootObj := filepath.Join(GOROOT, "pkg", GOOS+"_"+GOARCH)

	cdefargv := []string{GetCCompilerName(), "-FVw", "-I" + gorootObj}

	for _, objdst := range GOPATH_OBJDSTS {
		cdefargv = append(cdefargv, "-I"+objdst)
	}

	cdefargv = append(cdefargv, filepath.Join("_obj", "_cgo_defun.c"))

	if Verbose {
		fmt.Printf("%s:", cgodir)
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
	gccCompile := func(src, obj string) (err error) {
		gccargv := []string{"gcc", "-I..", "-I."}
		gccargv = append(gccargv, CFLAGS...)
		gccargv = append(gccargv, []string{"-g", "-fPIC", "-O2", "-o", obj, "-c"}...)
		gccargv = append(gccargv, pkg.CGoCFlags[pkg.Name]...)
		gccargv = append(gccargv, src)
		if Verbose {
			fmt.Printf("%s:", cgodir)
		}
		err = RunExternal(GCCCMD, cgodir, gccargv)
		return
	}
	var cobjs []string
	for _, cgb := range cgobases {
		cgc := cgb[:len(cgb)-3] + ".cgo2.c"
		cgo := cgb[:len(cgb)-3] + ".cgo2.o"
		cobjs = append(cobjs, cgo)

		src := filepath.Join("_obj", cgc)

		err = gccCompile(src, cgo)
		if err != nil {
			return
		}
	}

	for _, csrc := range pkg.CSrcs {
		cobj := csrc[:len(csrc)-2] + ".o"
		cobj = filepath.Base(cobj)
		cobjs = append(cobjs, cobj)
		relsrc := GetRelative("_cgo", csrc, filepath.Join(CWD, pkg.Dir))
		err = gccCompile(relsrc, cobj)
		if err != nil {
			return
		}
	}

	if err = gccCompile(filepath.Join("_obj", "_cgo_export.c"), "_cgo_export.o"); err != nil {
		return
	}
	cobjs = append(cobjs, "_cgo_export.o")
	if err = gccCompile(filepath.Join("_obj", "_cgo_main.c"), "_cgo_main.o"); err != nil {
		return
	}

	/* and link them
	gcc -m64 -g -fPIC -O2 -o _cgo1_.o _cgo_main.o e1.cgo2.o e2.cgo2.o _cgo_export.o
	*/
	gcclargv := []string{"gcc"}
	gcclargv = append(gcclargv, CFLAGS...)
	gcclargv = append(gcclargv, []string{"-g", "-fPIC", "-O2", "-o", "_cgo1_.o"}...)
	gcclargv = append(gcclargv, "_cgo_main.o")
	gcclargv = append(gcclargv, cobjs...)
	gcclargv = append(gcclargv, pkg.CGoLDFlags[pkg.Name]...)

	if Verbose {
		fmt.Printf("%s:", cgodir)
	}
	err = RunExternal(GCCCMD, cgodir, gcclargv)
	if err != nil {
		return
	}

	//cgo -dynimport _cgo1_.o >__cgo_import.c && mv -f __cgo_import.c _cgo_import.c
	dynargv := []string{"cgo", "-dynimport", "_cgo1_.o"}
	if Verbose {
		fmt.Printf("%s:", cgodir)
		fmt.Printf("writing to %s\n", "__cgo_import.c")
	}

	var dump *os.File
	dump, err = os.Create(filepath.Join(cgodir, "__cgo_import.c"))
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
	err = os.Rename(filepath.Join(cgodir, "__cgo_import.c"), filepath.Join(cgodir, "_cgo_import.c"))
	if err != nil {
		return
	}

	/* compile the C bits
	6c -FVw _cgo_import.c
	*/
	ccargv := []string{GetCCompilerName(), "-FVw", "_cgo_import.c"}
	if Verbose {
		fmt.Printf("%s:", cgodir)
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
	dstDir, _ := filepath.Split(dst)
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
		relobjs = append(relobjs, filepath.Join("_cgo", cobj))
	}
	packargv := []string{"gopack", "grc", reldst, GetIBName(),
		filepath.Join("_cgo", "_cgo_defun"+GetObjSuffix()),
		filepath.Join("_cgo", "_cgo_import"+GetObjSuffix())}
	packargv = append(packargv, relobjs...)

	err = RunExternal(PackCMD, pkg.Dir, packargv)
	return
}

func CleanCGoPackage(pkg *Package) (err error) {
	if !TestCGO {
		err = MakeClean(pkg)
		return
	}

	if Verbose {
		fmt.Printf(" Removing %s\n", filepath.Join(pkg.Dir, "_cgo"))
	}
	os.RemoveAll(filepath.Join(pkg.Dir, "_cgo"))

	return
}
