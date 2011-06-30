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
	//"time"
	"fmt"
	"os"
	"path"
)

func CompilePkgSrc(pkg *Package, src []string, obj, pkgDest string) (err os.Error) {

	argv := []string{GetCompilerName()}
	if len(GCFLAGS) > 0 {
		argv = append(argv, GCFLAGS...)
	}
	if !pkg.IsInGOROOT {
		argv = append(argv, "-I", pkgDest)
	}
	argv = append(argv, "-o", obj)
	argv = append(argv, src...)
	if Verbose {
		fmt.Printf("%v\n", argv)
	}
	err = RunExternal(CompileCMD, pkg.Dir, argv)
	return

}

func BuildPackage(pkg *Package) (err os.Error) {
	buildBlock <- true
	defer func() { 
		<-buildBlock
	}()

	pkgDest := GetRelative(pkg.Dir, GetBuildDirPkg(), CWD)

	err = CompilePkgSrc(pkg, pkg.PkgSrc[pkg.Name], GetIBName(), pkgDest)

	if err != nil {
		return
	}

	asmObjs := []string{}
	for _, asm := range pkg.AsmSrcs {
		base := asm[0 : len(asm)-2] // definitely ends with '.s', so this is safe
		asmObj := base + GetObjSuffix()
		asmObjs = append(asmObjs, asmObj)
		sargv := []string{GetAssemblerName(), asm}
		if Verbose {
			fmt.Printf("%v\n", sargv)
		}
		err = RunExternal(AsmCMD, pkg.Dir, sargv)
		if err != nil {
			return
		}
	}
	
	dst := GetRelative(pkg.Dir, pkg.ResultPath, CWD)

	if pkg.IsCmd {

		largs := []string{GetLinkerName()}
		
		if len(GLDFLAGS) > 0 {
			largs = append(largs, GLDFLAGS...)
		}
		
		if !pkg.IsInGOROOT {
			largs = append(largs, "-L", pkgDest)
		}
		
		//largs = append(largs, "-o", dst, GetIBName())
		largs = append(largs, "-o", pkg.Target, GetIBName())
		if Verbose {
			fmt.Printf("%v\n", largs)
		}
		//startLink := time.Nanoseconds()
		err = RunExternal(LinkCMD, pkg.Dir, largs)
		//durLink := time.Nanoseconds()-startLink
		//fmt.Printf("link took %f\n", float64(durLink)/1e9)
		os.MkdirAll(GetBuildDirCmd(), 0755)
		Copy(pkg.Dir, pkg.Target, dst)
	} else {
		dstDir, _ := path.Split(pkg.ResultPath)
		if Verbose {
			fmt.Printf("Creating directory %s\n", dstDir)
		}
		os.MkdirAll(dstDir, 0755)

		argv := []string{"gopack", "grc", dst, GetIBName()}
		argv = append(argv, asmObjs...)
		if Verbose {
			fmt.Printf("%v\n", argv)
		}
		if err = RunExternal(PackCMD, pkg.Dir, argv); err != nil {
			return
		}
	}

	var resInfo *os.FileInfo
	resInfo, err2 := os.Stat(pkg.ResultPath)
	if err2 == nil {
		pkg.BinTime = resInfo.Mtime_ns
	}

	return
}
func BuildTest(pkg *Package) (err os.Error) {

	reverseDots := ReverseDir(pkg.Dir)
	pkgDest := path.Join(reverseDots, GetBuildDirPkg())

	testIB := path.Join("_test", "_gotest_"+GetObjSuffix())

	//fmt.Printf("%v %v\n", pkg.TestSrc, pkg.Name)

	for testName, testSrcs := range pkg.TestSrc {

		argv := []string{GetCompilerName()}
		if GCFLAGS != nil {
			argv = append(argv, GCFLAGS...)
		}
		argv = append(argv, "-I", pkgDest)
		argv = append(argv, "-o", testIB)
		if testName == pkg.Name {
			argv = append(argv, pkg.PkgSrc[pkg.Name]...)
		}
		argv = append(argv, testSrcs...)

		if Verbose {
			fmt.Printf("%v\n", argv)
		}
		if err = RunExternal(CompileCMD, pkg.Dir, argv); err != nil {
			return
		}

		//see if it was created
		if _, err = os.Stat(path.Join(pkg.Dir, testIB)); err != nil {
			return os.NewError("compile error")
		}

		dst := path.Join("_test", "_obj", "_test", testName) + ".a"

		mkdirdst := path.Join(pkg.Dir, "_test", "_obj", "_test", testName) + ".a"
		dstDir, _ := path.Split(mkdirdst)
		os.MkdirAll(dstDir, 0755)

		argv = []string{"gopack", "grc", dst, testIB}
		if Verbose {
			fmt.Printf("%v\n", argv)
		}
		if err = RunExternal(PackCMD, pkg.Dir, argv); err != nil {
			return
		}

	}

	testmainib := path.Join("_test", "_testmain"+GetObjSuffix())

	argv := []string{GetCompilerName()}
	if GCFLAGS != nil {
		argv = append(argv, GCFLAGS...)
	}
	argv = append(argv, "-I", path.Join("_test", "_obj"))
	argv = append(argv, "-I", pkgDest)
	argv = append(argv, "-o", testmainib)
	argv = append(argv, path.Join("_test", "_testmain.go"))

	if Verbose {
		fmt.Printf("%v\n", argv)
	}
	if err = RunExternal(CompileCMD, pkg.Dir, argv); err != nil {
		return
	}

	testBinary := path.Join("_test", "_testmain")
	if GOOS == "windows" {
		testBinary += ".exe"
	}

	largs := []string{GetLinkerName()}
	if len(GLDFLAGS) > 0 {
		largs = append(largs, GLDFLAGS...)
	}
	largs = append(largs, "-L", path.Join("_test", "_obj"))
	largs = append(largs, "-L", pkgDest)
	largs = append(largs, "-o", testBinary, testmainib)
	if Verbose {
		fmt.Printf("%v\n", largs)
	}
	if err = RunExternal(LinkCMD, pkg.Dir, largs); err != nil {
		return
	}
	var testBinaryAbs string
	testBinaryAbs = GetAbs(path.Join(pkg.Dir, testBinary), CWD)
	if err = RunExternal(testBinaryAbs, pkg.Dir, []string{testBinary}); err != nil {
		ReturnFailCode = true
		return
	}

	return
}
func InstallPackage(pkg *Package) (err os.Error) {
	dstDir, _ := path.Split(pkg.InstallPath)
	_, dstName := path.Split(pkg.ResultPath)
	dstFile := path.Join(dstDir, dstName)
	err = os.MkdirAll(dstDir, 0755)
	if err != nil {
		return
	}

	which := "cmd"
	if pkg.Name != "main" {
		which = "pkg"
	}
	fmt.Printf("Installing %s \"%s\"\n", which, pkg.Target)

	Copy(".", pkg.ResultPath, dstFile)

	return
}
