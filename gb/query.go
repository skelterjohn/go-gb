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
	"runtime"
	"strings"
)

var GOROOT, GOOS, GOARCH, GOBIN string
var OSWD, CWD string

var GCFLAGS, GLDFLAGS []string

var GOPATH, GOPATH_SINGLE string
var GOPATHS, GOPATH_SRCROOTS, GOPATH_OBJDSTS, GOPATH_CFLAGS, GOPATH_LDFLAGS []string

func LoadCWD() (err error) {
	var oserr error
	OSWD, oserr = os.Getwd()

	CWD, err = OSWD, oserr

	runningInGOPATH := false

	if GOPATH = os.Getenv("GOPATH"); GOPATH != "" {
		gopaths := filepath.SplitList(GOPATH)
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
					fmt.Printf("Running gb in GOPATH workspace %s\n", CWD)
					runningInGOPATH = true
				}
			}
		}
	}

	if !runningInGOPATH {
		cfg := ReadConfig(".")
		if rel, set := cfg.Workspace(); set {
			CWD = GetAbs(filepath.Join(OSWD, rel), OSWD)
			fmt.Printf("Running gb in workspace %s\n", CWD)
		}
	}
	os.Chdir(CWD)

	return
}

func LoadEnvs() bool {
	GOROOT = runtime.GOROOT()

	GOOS, GOARCH, GOBIN = os.Getenv("GOOS"), os.Getenv("GOARCH"), os.Getenv("GOBIN")
	if GOOS == "" {
		GOOS = runtime.GOOS
		os.Setenv("GOOS", GOOS)
	}
	if GOARCH == "" {
		GOARCH = runtime.GOARCH
		os.Setenv("GOARCH", GOARCH)
	}
	if GOBIN == "" {
		GOBIN = filepath.Join(GOROOT, "bin")
		os.Setenv("GOBIN", GOBIN)
	}

	if !arch_flags[GOARCH] {
		ErrLog.Printf("Unknown GOARCH %s", GOARCH)
		return false
	}

	if !os_flags[GOOS] {
		ErrLog.Printf("Unknown GOOS %s", GOOS)
		return false
	}

	GOPATH = os.Getenv("GOPATH")

	if GOPATH != "" {
		gopaths := filepath.SplitList(GOPATH)
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

	gcFlagsStr, gldFlagsStr := os.Getenv("GCFLAGS"), os.Getenv("GB_GLDFLAGS")
	if gcFlagsStr != "" {
		GCFLAGS = append(GCFLAGS, strings.Fields(gcFlagsStr)...)
	}
	if gldFlagsStr != "" {
		GLDFLAGS = append(GLDFLAGS, strings.Fields(gldFlagsStr)...)
	}

	GCFLAGS = append(GCFLAGS, GOPATH_CFLAGS...)
	GLDFLAGS = append(GLDFLAGS, GOPATH_LDFLAGS...)

	RunningInGOROOT = HasPathPrefix(CWD, filepath.Join(GOROOT, "src"))

	return true
}

func GetBuildDirPkg() (dir string) {
	return ObjDir
}

func GetGOROOTDirPkg() (dir string) {
	return filepath.Join(GOROOT, "pkg", GOOS+"_"+GOARCH)
}

func GetInstallDirPkg() (dir string) {
	if GOPATH_SINGLE != "" {
		return filepath.Join(GOPATH_SINGLE, "pkg", GOOS+"_"+GOARCH)
	}
	return GetGOROOTDirPkg()
}

func GetBuildDirCmd() (dir string) {
	return BinDir
}

func GetInstallDirCmd() (dir string) {
	if GOPATH_SINGLE != "" {
		return filepath.Join(GOPATH_SINGLE, "bin")
	}
	return GOBIN
}

func ArchChar() (c string) {
	switch GOARCH {
	case "amd64":
		return "6"
	case "386":
		return "8"
	case "arm":
		return "5"
	}
	panic("unknown arch " + GOARCH)
	return
}

func GetCompilerName() (name string) {
	return ArchChar() + "g"
}

func GetCCompilerName() (name string) {
	return ArchChar() + "c"
}

func GetAssemblerName() (name string) {
	return ArchChar() + "a"
}

func GetLinkerName() (name string) {
	return ArchChar() + "l"
}

func GetObjSuffix() (suffix string) {
	return "." + ArchChar()
}

func GetIBName() (name string) {
	return fmt.Sprintf("_go_%s", GetObjSuffix())
}
