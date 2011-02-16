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
	"path"
	"os"
	"strings"
	"bufio"
	"fmt"
)

func FilterFlag(src string) bool {
	os_flags := []string{"windows", "darwin", "freebsd", "linux"}
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
	if strings.Contains(src, "_bsd") &&
		!(GOOS == "darwin" || GOOS == "freebsd" || GOOS == "bsd") {
		return false
	}

	return true
}

func GetSubDirs(dir string) (subdirs []string) {
	file, err := os.Open(dir, os.O_RDONLY, 0)
	if err != nil {
		return
	}
	infos, err := file.Readdir(-1)
	if err != nil {
		return
	}
	for _, info := range infos {
		if info.IsDirectory() {
			subdirs = append(subdirs, info.Name)
		}
	}
	return
}

func PkgExistsInGOROOT(target string) (exists bool, time int64) {
	if target[0] == '"' {
		target = target[1:len(target)]
	}
	if target[len(target)-1] == '"' {
		target = target[0 : len(target)-1]
	}

	pkgbin := path.Join(GetInstallDirPkg(), target)
	pkgbin += ".a"

	time, err := StatTime(pkgbin)

	exists = err == nil

	return
}

func LineChan(f string, ch chan<- string) (err os.Error) {
	var fin *os.File
	if fin, err = os.Open(f, os.O_RDONLY, 0); err == nil {
		bfrd := bufio.NewReader(fin)
		for {
			var line string
			if line, err = bfrd.ReadString('\n'); err != nil {
				break
			}
			ch <- strings.TrimSpace(line)
		}
	}
	return
}

func ReadOneLine(file string) (line string, err os.Error) {
	var fin *os.File
	fin, err = os.Open(file, os.O_RDONLY, 0)
	if err == nil {
		bfrd := bufio.NewReader(fin)
		line, err = bfrd.ReadString('\n')
		line = strings.TrimSpace(line)
	}
	return
}

func DirTargetGB(dir string) (target string, err os.Error) {
	target, err = ReadOneLine(path.Join(dir, "target.gb"))
	return
}


func ReverseDir(dir string) (rev string) {
	rev = "."
	for dir != "." && dir != "" {
		dir, _ = path.Split(path.Clean(dir))
		rev = path.Join(rev, "..")
	}
	return
}

func ReverseDirForwardSlash(dir string) (rev string) {
	rev = "."
	for dir != "." && dir != "" {
		dir, _ = path.Split(path.Clean(dir))
		rev += "/.."
	}
	return
}

func StatTime(p string) (time int64, err os.Error) {
	var info *os.FileInfo
	info, err = os.Stat(p)
	if err != nil {
		return
	}
	time = info.Mtime_ns
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
		var n int
		n, cpErr = srcFile.Read(buffer)
		if cpErr != nil {
			break
		}
		_, cpErr = dstFile.Write(buffer[0:n])
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
	if err = RunExternal(CopyCMD, cwd, argv); err != nil {
		return
	}

	return
}
