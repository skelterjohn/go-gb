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
	//"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func CompilePkgSrc(pkg *Package, src []string, obj, pkgDest, testDest string) (err error) {

	argv := []string{GetCompilerName()}
	if !pkg.IsInGOROOT {
		argv = append(argv, "-I", pkgDest)
	}
	if testDest != "" {
		argv = append(argv, "-I", testDest)
	}
	if len(GCFLAGS) > 0 {
		argv = append(argv, GCFLAGS...)
	}
	argv = append(argv, "-o", obj)
	if gcflags, set := pkg.Cfg.GCFlags(); set {
		argv = append(argv, strings.Fields(gcflags)...)
	}
	argv = append(argv, src...)

	err = RunExternal(CompileCMD, pkg.Dir, argv)
	return

}

func BuildPackage(pkg *Package) (err error) {

	pkgDest := GetRelative(pkg.Dir, GetBuildDirPkg(), CWD)

	var testDest string
	if pkg.InTestData != "" {
		tdBuildDir := filepath.Join(pkg.InTestData, GetBuildDirPkg())
		testDest = GetRelative(pkg.Dir, tdBuildDir, CWD)
	}

	ibname := GetIBName()

	err = CompilePkgSrc(pkg, pkg.PkgSrc[pkg.Name], ibname, pkgDest, testDest)

	if err != nil {
		return
	}

	defer func() {
		if Verbose {
			fmt.Printf("Removing %s\n", filepath.Join(pkg.Dir, ibname))
		}
		os.Remove(filepath.Join(pkg.Dir, ibname))
	}()

	asmObjs := []string{}
	for _, asm := range pkg.AsmSrcs {
		base := asm[0 : len(asm)-2] // definitely ends with '.s', so this is safe
		asmObj := base + GetObjSuffix()
		asmObjs = append(asmObjs, asmObj)
		sargv := []string{GetAssemblerName(), asm}

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
		if testDest != "" {
			largs = append(largs, "-L", testDest)
		}

		//largs = append(largs, "-o", dst, GetIBName())
		largs = append(largs, "-o", pkg.Target, GetIBName())

		//startLink := time.Nanoseconds()
		err = RunExternal(LinkCMD, pkg.Dir, largs)
		//durLink := time.Nanoseconds()-startLink
		//fmt.Printf("link took %f\n", float64(durLink)/1e9)
		dstDir, _ := filepath.Split(pkg.ResultPath)
		if Verbose {
			fmt.Printf("Creating directory %s\n", dstDir)
		}
		os.MkdirAll(dstDir, 0755)
		Copy(pkg.Dir, pkg.Target, dst)
	} else {
		dstDir, _ := filepath.Split(pkg.ResultPath)
		if Verbose {
			fmt.Printf("Creating directory %s\n", dstDir)
		}
		os.MkdirAll(dstDir, 0755)

		argv := []string{"gopack", "grc", dst, GetIBName()}
		argv = append(argv, asmObjs...)

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
func BuildTest(pkg *Package) (err error) {

	reverseDots := ReverseDir(pkg.Dir)
	pkgDest := filepath.Join(reverseDots, GetBuildDirPkg())

	testIB := filepath.Join("_test", "_gotest_"+GetObjSuffix())

	//fmt.Printf("%v %v\n", pkg.TestSrc, pkg.Name)

	buildTestName := func(testName string) (err error) {

		testSrcs := pkg.TestSrc[testName]

		argv := []string{GetCompilerName()}
		argv = append(argv, "-I", filepath.Join("_test", "_obj"))
		argv = append(argv, "-I", pkgDest)
		if GCFLAGS != nil {
			argv = append(argv, GCFLAGS...)
		}
		argv = append(argv, "-o", testIB)
		if testName == pkg.Name {
			argv = append(argv, pkg.PkgSrc[pkg.Name]...)
		}
		argv = append(argv, testSrcs...)

		if err = RunExternal(CompileCMD, pkg.Dir, argv); err != nil {
			return
		}

		//see if it was created
		if _, err = os.Stat(filepath.Join(pkg.Dir, testIB)); err != nil {
			return errors.New("compile error")
		}
		dst := filepath.Join("_test", "_obj", testName) + ".a"

		if testName == pkg.Name {
			dst = filepath.Join("_test", "_obj", pkg.Target) + ".a"
		}

		mkdirdst := filepath.Join(pkg.Dir, dst)
		dstDir, _ := filepath.Split(mkdirdst)
		os.MkdirAll(dstDir, 0755)

		argv = []string{"gopack", "grc", dst, testIB}

		if err = RunExternal(PackCMD, pkg.Dir, argv); err != nil {
			return
		}

		return
	}

	err = buildTestName(pkg.Name)
	if err != nil {
		return
	}

	for testName := range pkg.TestSrc {
		if testName == pkg.Name {
			continue
		}
		err = buildTestName(testName)
		if err != nil {
			return
		}
	}

	testmainib := filepath.Join("_test", "_testmain"+GetObjSuffix())

	argv := []string{GetCompilerName()}
	argv = append(argv, "-I", filepath.Join("_test", "_obj"))
	argv = append(argv, "-I", pkgDest)
	if GCFLAGS != nil {
		argv = append(argv, GCFLAGS...)
	}
	argv = append(argv, "-o", testmainib)
	argv = append(argv, filepath.Join("_test", "_testmain.go"))

	if err = RunExternal(CompileCMD, pkg.Dir, argv); err != nil {
		return
	}

	testBinary := filepath.Join("_test", "_testmain")
	if GOOS == "windows" {
		testBinary += ".exe"
	}

	largs := []string{GetLinkerName()}
	largs = append(largs, "-L", filepath.Join("_test", "_obj"))
	largs = append(largs, "-L", pkgDest)
	if len(GLDFLAGS) > 0 {
		largs = append(largs, GLDFLAGS...)
	}
	largs = append(largs, "-o", testBinary, testmainib)

	if err = RunExternal(LinkCMD, pkg.Dir, largs); err != nil {
		return
	}
	var testBinaryAbs string
	testBinaryAbs = GetAbs(filepath.Join(pkg.Dir, testBinary), CWD)
	testargs := append([]string{testBinary}, TestArgs...)

	if err = RunExternal(testBinaryAbs, pkg.Dir, testargs); err != nil {
		ReturnFailCode = true
		return
	}

	return
}
func InstallPackage(pkg *Package) (err error) {
	dstDir, _ := filepath.Split(pkg.InstallPath)
	_, dstName := filepath.Split(pkg.ResultPath)
	dstFile := filepath.Join(dstDir, dstName)
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
