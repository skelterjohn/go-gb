/* 
* Copyright (C) 2011, John Asmuth

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
* 
*  0.7
*  1/12/2001
*  John Asmuth
*  http://go-gb.googlecode.com
* 
*/

package main

import (
	"fmt"
	"os"
	"exec"
	"path"
)


func GetCompilerName() (name string) {
	switch GOARCH {
	case "amd64":
		return "6g"
	case "386":
		return "8g"
	case "arm":
		return "5g"
	}
	return
}

func GetLinkerName() (name string) {
	switch GOARCH {
	case "amd64":
		return "6l"
	case "386":
		return "8l"
	case "arm":
		return "5l"
	}
	return
}

func GetObjSuffix() (suffix string) {
	switch GOARCH {
	case "amd64":
		return ".6"
	case "386":
		return ".5"
	case "arm":
		return ".8"
	}
	return
}

func GetIBName() (name string) {
	return "_go_" + GetObjSuffix()
}

func ReverseDir(dir string) (rev string) {
	rev = "."
	for dir != "." && dir != "" {
		dir, _ = path.Split(path.Clean(dir))
		rev = path.Join(rev, "..")
	}
	return
}

func BuildPackage(pkg *Package) (err os.Error) {
	/*
		relativeSources := make([]string, len(pkg.Sources))
		for i, src := range pkg.Sources {
			relativeSources[i] = path.Join(pkg.Dir, src)
		}
	*/

	reverseDots := ReverseDir(pkg.Dir)
	pkgDest := path.Join(reverseDots, "_obj")

	argv := []string{GetCompilerName()}
	argv = append(argv, "-I", pkgDest)
	argv = append(argv, "-o", GetIBName())
	argv = append(argv, pkg.Sources...)
	if Verbose {
		fmt.Printf("%v\n", argv)
	}
	p, err := exec.Run(CompileCMD, argv, os.Envs, pkg.Dir, exec.PassThrough, exec.PassThrough, exec.PassThrough)
	if err != nil {
		return
	}
	if p != nil {
		p.Wait(0)
	}

	//see if it was created
	if _, err = os.Stat(pkg.ib); err != nil {
		return os.NewError("compile error")
	}

	if pkg.IsCmd {
		dst := pkg.Target
		largs := []string{GetLinkerName()}
		largs = append(largs, "-L", pkgDest)
		largs = append(largs, "-o", dst, GetIBName())
		if Verbose {
			fmt.Printf("%v\n", largs)
		}
		p, err = exec.Run(LinkCMD, largs, os.Envs, pkg.Dir, exec.PassThrough, exec.PassThrough, exec.PassThrough)
		if err != nil {
			return
		}
		if p != nil {
			p.Wait(0)
		}
	} else {
		dst := path.Join(pkgDest, pkg.Target) + ".a"

		mkdirdst := path.Join("_obj", pkg.Target) + ".a"
		dstDir, _ := path.Split(mkdirdst)
		os.MkdirAll(dstDir, 0755)

		argv = []string{"gopack", "grc", dst, GetIBName()}
		if Verbose {
			fmt.Printf("%v\n", argv)
		}
		p, err = exec.Run(PackCMD, argv, os.Envs, pkg.Dir, exec.PassThrough, exec.PassThrough, exec.PassThrough)
		if err != nil {
			return
		}
		if p != nil {
			p.Wait(0)
		}
	}

	var resInfo *os.FileInfo
	resInfo, err2 := os.Stat(pkg.result)
	if err2 == nil {
		pkg.BinTime = resInfo.Mtime_ns
	}

	return
}
func BuildTest(pkg *Package) (err os.Error) {

	reverseDots := ReverseDir(pkg.Dir)
	pkgDest := path.Join(reverseDots, "_obj")

	testIB := path.Join("_test", "_gotest_"+GetObjSuffix())

	argv := []string{GetCompilerName()}
	argv = append(argv, "-I", pkgDest)
	argv = append(argv, "-o", testIB)
	argv = append(argv, pkg.Sources...)
	argv = append(argv, pkg.TestSources...)

	if Verbose {
		fmt.Printf("%v\n", argv)
	}
	p, err := exec.Run(CompileCMD, argv, os.Envs, pkg.Dir, exec.PassThrough, exec.PassThrough, exec.PassThrough)
	if err != nil {
		return
	}
	if p != nil {
		p.Wait(0)
	}

	//see if it was created
	if _, err = os.Stat(path.Join(pkg.Dir, testIB)); err != nil {
		return os.NewError("compile error")
	}

	dst := path.Join("_test", "_obj", pkg.Target) + ".a"

	mkdirdst := path.Join(pkg.Dir, "_test", "_obj", pkg.Target) + ".a"
	dstDir, _ := path.Split(mkdirdst)
	os.MkdirAll(dstDir, 0755)

	argv = []string{"gopack", "grc", dst, testIB}
	if Verbose {
		fmt.Printf("%v\n", argv)
	}
	p, err = exec.Run(PackCMD, argv, os.Envs, pkg.Dir, exec.PassThrough, exec.PassThrough, exec.PassThrough)
	if err != nil {
		return
	}
	if p != nil {
		p.Wait(0)
	}

	testmainib := path.Join("_test", "_testmain"+GetObjSuffix())

	argv = []string{GetCompilerName()}
	argv = append(argv, "-I", path.Join("_test", "_obj"))
	argv = append(argv, "-I", pkgDest)
	argv = append(argv, "-o", testmainib)
	argv = append(argv, path.Join("_test", "_testmain.go"))

	if Verbose {
		fmt.Printf("%v\n", argv)
	}
	p, err = exec.Run(CompileCMD, argv, os.Envs, pkg.Dir, exec.PassThrough, exec.PassThrough, exec.PassThrough)
	if err != nil {
		return
	}
	if p != nil {
		p.Wait(0)
	}

	testBinary := "_testmain"
	if GOOS == "windows" {
		testBinary += ".exe"
	}

	largs := []string{GetLinkerName()}
	largs = append(largs, "-L", path.Join("_test", "_obj"))
	largs = append(largs, "-L", pkgDest)
	largs = append(largs, "-o", path.Join("_test", testBinary), testmainib)
	if Verbose {
		fmt.Printf("%v\n", largs)
	}
	p, err = exec.Run(LinkCMD, largs, os.Envs, pkg.Dir, exec.PassThrough, exec.PassThrough, exec.PassThrough)
	if err != nil {
		return
	}
	if p != nil {
		p.Wait(0)
	}

	if err != nil {
		return
	}

	p, err = exec.Run(testBinary, []string{testBinary}, os.Envs, path.Join(pkg.Dir, "_test"), exec.PassThrough, exec.PassThrough, exec.PassThrough)
	if err != nil {
		return
	}
	if p != nil {
		p.Wait(0)
	}

	return
}
func InstallPackage(pkg *Package) (err os.Error) {
	dstDir, _ := path.Split(pkg.installPath)
	_, dstName := path.Split(pkg.result)
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

	err = Copy(".", pkg.result, dstFile)
	/*
	argv := append([]string{"cp", "-f", pkg.result, dstDir})
	if Verbose {
		fmt.Printf("%v\n", argv)
	}
	p, err := exec.Run(CopyCMD, argv, os.Envs, ".", exec.PassThrough, exec.PassThrough, exec.PassThrough)
	if err != nil {
		return
	}
	if p != nil {
		p.Wait(0)
	}
	*/
	return
}

func CopyTheHardWay(cwd, src, dst string) (err os.Error) {
	srcpath := path.Join(cwd, src)
	
	if Verbose {
		fmt.Printf("Copying %s to %s\n", src, dst)
	}
	
	dstpath := dst
	if !path.IsAbs(dstpath) {
		dstpath = path.Join(cwd, dst)
	}
	
	var srcFile *os.File
	srcFile, err = os.Open(srcpath, os.O_RDONLY, 0)
	if err != nil {
		return
	}
	
	var dstFile *os.File
	dstFile, err = os.Open(dstpath, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return
	}
	
	buffer := make([]byte, 1024)
	var cpErr os.Error
	for {
		_, cpErr = srcFile.Read(buffer)
		if cpErr != nil {
			break
		}
		_, cpErr = dstFile.Write(buffer)
		if cpErr != nil {
			break
		}
	}
	if cpErr != os.EOF {
		err = cpErr
	}

	dstFile.Close()
	
	return
}

func Copy(cwd, src, dst string) (err os.Error) {
	if CopyCMD == "" {
		return CopyTheHardWay(cwd, src, dst)
	}
	
	argv := append([]string{"cp", "-f", src, dst})
	if Verbose {
		fmt.Printf("%v\n", argv)
	}
	p, err := exec.Run(CopyCMD, argv, os.Envs, cwd, exec.PassThrough, exec.PassThrough, exec.PassThrough)
	if err != nil {
		return err
	}
	if p != nil {
		p.Wait(0)
	}
	
	return
}
