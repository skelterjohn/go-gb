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
	"strconv"
)

var GOROOT, GOOS, GOARCH, GOBIN string
var CWD string

func LoadEnvs() {

	GOOS, GOARCH, GOROOT, GOBIN = os.Getenv("GOOS"), os.Getenv("GOARCH"), os.Getenv("GOROOT"), os.Getenv("GOBIN")
	if GOOS == "" {
		println("Environental variable GOOS not set")
		return
	}
	if GOARCH == "" {
		println("Environental variable GOARCH not set")
		return
	}
	if GOROOT == "" {
		println("Environental variable GOROOT not set")
		return
	}
	if GOBIN == "" {
		GOBIN = path.Join(GOROOT, "bin")
	}

	CWD, _ = os.Getwd()
	RunningInGOROOT = strings.HasPrefix(CWD, GOROOT)

	GOMAXPROCS := os.Getenv("GOMAXPROCS")

	n, nerr := strconv.Atoi(GOMAXPROCS)
	if nerr != nil {
		n = 1
	}
	n = 2
	buildBlock = make(chan bool, n)
}

func GetBuildDirPkg() (dir string) {
	return "_obj"
}

func GetInstallDirPkg() (dir string) {
	return path.Join(GOROOT, "pkg", GOOS+"_"+GOARCH)
}

func GetBuildDirCmd() (dir string) {
	return "bin"
}

func GetInstallDirCmd() (dir string) {
	return path.Join(GOROOT, "bin")
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
