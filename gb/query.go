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
	"strings"
	"runtime"
)

var GOROOT, GOOS, GOARCH, GOBIN string
var OSWD, CWD string

var GCFLAGS, GLDFLAGS []string

var GOPATH, GOPATH_SINGLE string
var GOPATHS, GOPATH_SRCROOTS, GOPATH_OBJDSTS, GOPATH_CFLAGS, GOPATH_LDFLAGS []string


func LoadCWD() (err os.Error) {
	var oserr os.Error
	OSWD, oserr = os.Getwd()
	rel, relerr := ReadOneLine("workspace.gb")

	CWD, err = OSWD, oserr

	if relerr == nil {
		CWD = GetAbs(filepath.Join(OSWD, rel), OSWD)
		fmt.Printf("Running gb in workspace %s\n", CWD)
	} else if GOPATH=os.Getenv("GOPATH"); GOPATH != "" {
		gopaths := strings.Split(GOPATH, ":", -1)
		if GOOS == "windows" {
			gopaths = strings.Split(GOPATH, ";", -1)
		}
		for _, gp := range gopaths {
			gp = strings.TrimSpace(gp)
			if gp == "" {
				continue
			}
			gpsrc := filepath.Join(gp, "src")

			if HasPathPrefix(OSWD, gp) {
				RunningInGOPATH = gp
				if CWD != gpsrc {
					CWD = gpsrc
					fmt.Printf("Running gb in GOPATH %s\n", CWD)
					os.Chdir(CWD)
				}
			}
		}
	}

	os.Chdir(CWD)
	return
}

func LoadEnvs() bool {

	GOOS, GOARCH, GOROOT, GOBIN = os.Getenv("GOOS"), os.Getenv("GOARCH"), os.Getenv("GOROOT"), os.Getenv("GOBIN")
	if GOOS == "" {
		GOOS = runtime.GOOS
		os.Setenv("GOOS", GOOS)
	}
	if GOARCH == "" {
		GOARCH = runtime.GOARCH
		os.Setenv("GOARCH", GOARCH)
	}
	if GOROOT == "" {
		ErrLog.Printf("Environental variable GOROOT not set")
		return false
	}
	if GOBIN == "" {
		GOBIN = filepath.Join(GOROOT, "bin")
		os.Setenv("GOBIN", GOBIN)
	}

	GOPATH = os.Getenv("GOPATH")
	
	if GOPATH != "" {
		gopaths := strings.Split(GOPATH, ":", -1)
		if GOOS == "windows" {
			gopaths = strings.Split(GOPATH, ";", -1)
		}
		for _, gp := range gopaths {
			gp = strings.TrimSpace(gp)
			if gp == "" {
				continue
			}

			gpsrc := filepath.Join(gp, "src")


			GOPATHS = append(GOPATHS, gp)

			if GOPATH_SINGLE == "" {
				GOPATH_SINGLE = gp
			}

			GOPATH_SRCROOTS = append(GOPATH_SRCROOTS, gpsrc)
			objdst := filepath.Join(gp, "pkg", fmt.Sprintf("%s_%s", GOOS, GOARCH))
			GOPATH_OBJDSTS = append(GOPATH_OBJDSTS, objdst)
			GOPATH_CFLAGS = append(GOPATH_CFLAGS, "-I", objdst)
			GOPATH_LDFLAGS = append(GOPATH_LDFLAGS, "-L", objdst)

			os.MkdirAll(objdst, 0755)
		}
	}

	gcFlagsStr, gldFlagsStr := os.Getenv("GB_GCFLAGS"), os.Getenv("GB_GLDFLAGS")
	if gcFlagsStr != "" {
		GCFLAGS = append(GCFLAGS, strings.Fields(gcFlagsStr)...)
	}
	if gldFlagsStr != "" {
		GLDFLAGS = append(GLDFLAGS, strings.Fields(gldFlagsStr)...)
	}

	GCFLAGS = append(GCFLAGS, GOPATH_CFLAGS...)
	GLDFLAGS = append(GLDFLAGS, GOPATH_LDFLAGS...)
	
	RunningInGOROOT = HasPathPrefix(CWD, filepath.Join(GOROOT, "src"))

	buildBlock = make(chan bool, runtime.GOMAXPROCS(0)) //0 doesn't change, only returns

	return true
}

func GetBuildDirPkg() (dir string) {
	return "_obj"
}

func GetInstallDirPkg() (dir string) {
	if GOPATH_SINGLE != "" {
		return filepath.Join(GOPATH_SINGLE, "pkg", GOOS+"_"+GOARCH)
	}
	return filepath.Join(GOROOT, "pkg", GOOS+"_"+GOARCH)
}

func GetBuildDirCmd() (dir string) {
	return "bin"
}

func GetInstallDirCmd() (dir string) {
	if GOPATH_SINGLE != "" {
		return filepath.Join(GOPATH_SINGLE, "bin")
	}
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
