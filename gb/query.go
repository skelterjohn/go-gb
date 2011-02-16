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
	"path"
	"strings"
	"runtime"
)

var GOROOT, GOOS, GOARCH, GOBIN string
var OSWD, CWD string

func LoadCWD() (err os.Error) {
	var oserr os.Error
	OSWD, oserr = os.Getwd()
	rel, relerr := ReadOneLine("workspace.gb")
	if relerr != nil {
		CWD, err = OSWD, oserr
	} else {
		CWD = GetAbs(path.Join(OSWD, rel), OSWD)
		fmt.Printf("Running gb in %s\n", CWD)
	}
	os.Chdir(CWD)
	return
}

func LoadEnvs() bool {

	GOOS, GOARCH, GOROOT, GOBIN = os.Getenv("GOOS"), os.Getenv("GOARCH"), os.Getenv("GOROOT"), os.Getenv("GOBIN")
	if GOOS == "" {
		println("Environental variable GOOS not set")
		return false
	}
	if GOARCH == "" {
		println("Environental variable GOARCH not set")
		return false
	}
	if GOROOT == "" {
		println("Environental variable GOROOT not set")
		return false
	}
	if GOBIN == "" {
		GOBIN = path.Join(GOROOT, "bin")
	}

	RunningInGOROOT = strings.HasPrefix(CWD, GOROOT)

	buildBlock = make(chan bool, runtime.GOMAXPROCS(0)) //0 doesn't change, only returns
	
	return true
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
	return GOBIN
}

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

func GetCCompilerName() (name string) {
	switch GOARCH {
	case "amd64":
		return "6c"
	case "386":
		return "8c"
	case "arm":
		return "5c"
	}
	return
}

func GetAssemblerName() (name string) {
	switch GOARCH {
	case "amd64":
		return "6a"
	case "386":
		return "8a"
	case "arm":
		return "5a"
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
		return ".8"
	case "arm":
		return ".5"
	}
	return
}

func GetIBName() (name string) {
	return "_go_" + GetObjSuffix()
}
