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
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

var (
	os_flags = map[string]bool{
		"windows": true,
		"darwin":  true,
		"freebsd": true,
		"openbsd": true,
		"linux":   true,
		"plan9":   true,
	}
	arch_flags = map[string]bool{
		"amd64": true,
		"386":   true,
		"arm":   true,
	}
)

func CheckCGOFlag(flag string) bool {
	if flag == GOOS || flag == GOARCH {
		return true
	}
	if flag == "unix" &&
		(GOOS == "darwin" || GOOS == "freebsd" || GOOS == "bsd" || GOOS == "linux") {
		return true
	}
	if flag == "posix" &&
		(GOOS == "darwin" || GOOS == "freebsd" || GOOS == "bsd" || GOOS == "linux" || GOOS == "windows") {
		return true
	}
	if flag == "bsd" &&
		(GOOS == "darwin" || GOOS == "freebsd" || GOOS == "bsd" || GOOS == "openbsd") {
		return true
	}
	return false
}

func FilterFlag(src string) bool {
	for flag := range os_flags {
		if strings.Contains(src, "_"+flag) && GOOS != flag {
			return false
		}
	}
	for flag := range arch_flags {
		if strings.Contains(src, "_"+flag) && GOARCH != flag {
			return false
		}
	}
	if strings.Contains(src, "_unix") &&
		!(GOOS == "darwin" || GOOS == "freebsd" || GOOS == "bsd" || GOOS == "linux") {
		return false
	}
	if strings.Contains(src, "_posix") &&
		!(GOOS == "darwin" || GOOS == "freebsd" || GOOS == "bsd" || GOOS == "linux" || GOOS == "windows") {
		return false
	}
	if strings.Contains(src, "_bsd") &&
		!(GOOS == "darwin" || GOOS == "freebsd" || GOOS == "bsd") {
		return false
	}

	return true
}

func splitPathAll(p string) (bits []string) {
	if p == "/" {
		return []string{}
	}
	dir, base := path.Split(p)
	if dir != "" {
		bits = append(splitPathAll(path.Clean(dir)), base)
	} else {
		bits = []string{base}
	}
	return
}

//GOOS and GOARCH excluded if they don't match GOOS and GOARCH
func FilterPkg(dir string) bool {
	splitdir := splitPathAll(dir)
	for _, flag := range splitdir {
		if os_flags[flag] && flag != GOOS {
			return false
		}
		if arch_flags[flag] && flag != GOARCH {
			return false
		}
		if flag == "unix" {
			if !(GOOS == "darwin" || GOOS == "freebsd" || GOOS == "bsd" || GOOS == "linux") {
				return false
			}
		}
		if flag == "posix" {
			if !(GOOS == "darwin" || GOOS == "freebsd" || GOOS == "bsd" || GOOS == "linux" || GOOS == "windows") {
				return false
			}
		}
		if flag == "bsd" {
			if !(GOOS == "darwin" || GOOS == "freebsd" || GOOS == "bsd") {
				return false
			}
		}
	}
	return true
}

func GetSubDirs(dir string) (subdirs []string) {
	file, err := os.Open(dir)
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

	pkgbin := path.Join(GetGOROOTDirPkg(), target)
	pkgbin += ".a"

	time, err := StatTime(pkgbin)

	exists = err == nil

	return
}

func LineChan(f string, ch chan<- string) (err error) {
	var fin *os.File
	if fin, err = os.Open(f); err == nil {
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

func ReadOneLine(file string) (line string, err error) {
	var fin *os.File
	fin, err = os.Open(file)
	if err == nil {
		bfrd := bufio.NewReader(fin)
		line, err = bfrd.ReadString('\n')
		line = strings.TrimSpace(line)
	}
	return
}

func DirTargetGB(dir string) (target string, err error) {
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

func StatTime(p string) (time int64, err error) {
	var info *os.FileInfo
	info, err = os.Stat(p)
	if err != nil {
		return
	}
	time = info.Mtime_ns
	return
}

func CopyTheHardWay(cwd, src, dst string) (err error) {
	srcpath := path.Join(cwd, src)

	if Verbose {
		fmt.Printf("Copying %s to %s\n", src, dst)
	}

	dstpath := dst
	if !path.IsAbs(dstpath) {
		dstpath = path.Join(cwd, dst)
	}

	var srcFile *os.File
	srcFile, err = os.Open(srcpath)
	if err != nil {
		return
	}

	var dstFile *os.File
	dstFile, err = os.Create(dstpath)
	if err != nil {
		return
	}

	io.Copy(dstFile, srcFile)

	dstFile.Close()
	srcFile.Close()

	return
}

func Copy(cwd, src, dst string) (err error) {
	if CopyCMD == "" {
		return CopyTheHardWay(cwd, src, dst)
	}

	argv := append([]string{"cp", "-f", src, dst})

	if err = RunExternal(CopyCMD, cwd, argv); err != nil {
		return
	}

	return
}
