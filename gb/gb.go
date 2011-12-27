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

//target: gb
package main

import (
	//"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// command line flags
var Install, //-i
	Clean, //-c
	Nuke, //-N
	Scan, //-sS
	ScanList, //-S
	ScanListFiles, //-L
	Test, //-t
	Exclusive, //-e
	BuildGOROOT, //-R
	GoInstall, //-gG
	GoInstallUpdate, //-G
	Concurrent, //-p
	Verbose, //-v
	GenMake, //--makefiles
	Build, //-b
	Force, //-f
	Makefiles, //-m
	GoFMT, //--gofmt
	GoFix, //--gofix
	DoPkgs, //-P
	DoCmds, //-C
	Distribution, //--dist
	Workspace bool //--workspace

var IncludeDir string
var GCArgs []string
var GLArgs []string
var PackagesBuilt int
var PackagesCleaned int
var PackagesInstalled int
var BrokenPackages int
var ListedTargets int
var ListedDirs, ValidatedDirs map[string]bool
var ListedPkgs []*Package

var HardArgs, BuildArgs int

var TestArgs []string

var BrokenMsg []string
var ReturnFailCode bool

var RunningInGOROOT bool
var RunningInGOPATH string

var Packages = make(map[string]*Package)

var ErrLog = log.New(os.Stderr, "gb error: ", 0)
var WarnLog = log.New(os.Stderr, "gb warning: ", 0)

/*
 gb doesn't know how to build these packages

 go/build has a source-generation step that uses make variables

 os has source generation

 syscall has files type_$(GOOS).go that aren't build, but can't reasonably be filtered

 crypto/tls has a file root_stub.go which is excluded
*/
var ForceMakePkgs = map[string]bool{
	//"math":       true,
	"go/build":   true,
	"os":         true,
	"os/user":    true,
	"net":        true,
	"hash/crc32": true,
	"syscall":    true,
	"runtime":    true,
	"crypto/tls": true,
	"godoc":      true,
}

var DoNotBuildGOROOT = map[string]bool{
	"src/cmd/5a":     true,
	"src/cmd/5c":     true,
	"src/cmd/5g":     true,
	"src/cmd/5l":     true,
	"src/cmd/6a":     true,
	"src/cmd/6c":     true,
	"src/cmd/6g":     true,
	"src/cmd/6l":     true,
	"src/cmd/8a":     true,
	"src/cmd/8c":     true,
	"src/cmd/8g":     true,
	"src/cmd/8l":     true,
	"src/cmd/cc":     true,
	"src/cmd/gc":     true,
	"src/cmd/cov":    true,
	"src/cmd/gopack": true,
	"src/cmd/nm":     true,
	"src/cmd/prof":   true,
}

const (
	ObjDir  = "_obj"
	TestDir = "_test"
	CGoDir  = "_cgo"
	BinDir  = "_bin"
)

var DisallowedSourceDirectories = map[string]bool{
	ObjDir:  true,
	TestDir: true,
	CGoDir:  true,
	BinDir:  true,
}

var OSFiltersMust = map[string]string{
	"wingui": "windows",
}

func ScanDirectory(base, dir string, inTestData string) (err2 error) {
	_, basedir := filepath.Split(dir)
	if DisallowedSourceDirectories[basedir] || (basedir != "." && strings.HasPrefix(basedir, ".")) {
		return
	}

	if basedir == "testdata" {
		// if gb isn't actually run from within here, ignore it all
		if !HasPathPrefix(OSWD, GetAbs(dir, CWD)) {
			return
		}
		// all stuff within is for testing
		inTestData = dir
		// and it starts from scratch with the target name
		base = "."
	}

	cfg := ReadConfig(dir)

	if Workspace {
		absdir := GetAbs(dir, CWD)
		relworkspace := GetRelative(absdir, CWD, CWD)

		cfg["workspace"] = relworkspace
		if err := cfg.Write(absdir); err != nil {
			ErrLog.Println(err)
		}
	}

	if ignoreAll, ok := cfg.IgnoreAll(); ignoreAll && ok {
		return
	}

	var err error

	var pkg *Package

	if ignore, ok := cfg.Ignore(); !(ignore && ok) {
		pkg, err = NewPackage(base, dir, inTestData, cfg)
		if err == nil {
			key := "\"" + pkg.Target + "\""
			if pkg.IsCmd {
				key += "-cmd"
			}
			if dup, exists := Packages[key]; exists {
				if GetAbs(dup.Dir, CWD) != GetAbs(pkg.Dir, CWD) {
					ErrLog.Printf("Duplicate target: %s\n in %s\n in %s\n", pkg.Target, dup.Dir, pkg.Dir)
				}
			} else {
				Packages[key] = pkg
			}
			base = pkg.Base
		} else {
			if tbase, terr := DirTargetGB(dir); terr == nil {
				base = tbase
			}
		}
	} else {
		fmt.Println(dir, "ignored")
	}

	subdirs := GetSubDirs(dir)
	for _, subdir := range subdirs {
		ScanDirectory(filepath.Join(base, subdir), filepath.Join(dir, subdir), inTestData)
	}

	return
}

func ValidateDir(name string) {
	if Exclusive {
		ValidatedDirs[name] = true
		return
	}
	for lt := range ListedDirs {
		rel := GetRelative(lt, name, CWD)
		if !HasPathPrefix(rel, "..") {
			ValidatedDirs[lt] = true
		}
	}
}

func IsListed(name string) bool {
	if ListedTargets == 0 {
		return true
	}
	if Exclusive {
		return ListedDirs[name]
	}

	for lt := range ListedDirs {
		rel := GetRelative(lt, name, CWD)
		if !HasPathPrefix(rel, "..") {
			return true
		}
	}
	return false
}

func TryScan() {
	if Scan {
		for _, pkg := range Packages {
			if pkg.IsInGOROOT && !RunningInGOROOT {
				continue
			}
			if pkg.IsInGOPATH != "" && RunningInGOPATH == "" {
				continue
			}
			if IsListed(pkg.Dir) {
				pkg.PrintScan()
			}
		}
		return
	}
}

func TryGoFMT() (err error) {
	if GoFMT {
		for _, pkg := range ListedPkgs {
			err = pkg.GoFMT()
			if err != nil {
				return
			}
		}
	}
	return
}

func TryGoFix() (err error) {
	if GoFix {
		for _, pkg := range ListedPkgs {
			err = pkg.GoFix()
			if err != nil {
				return
			}
		}
	}
	return
}

func TryGenMake() (err error) {
	if GenMake {
		_, ferr := os.Stat("build")

		genBuild := true

		if ferr == nil {
			if !Force {
				fmt.Printf("'build' exists; overwrite? (y/n) ")
				var answer string
				fmt.Scanf("%s", &answer)
				genBuild = answer == "y" || answer == "Y"
			}
			os.Remove("build")
		}

		if genBuild {
			fmt.Printf("(in .) generating build script\n")
			var buildFile *os.File
			buildFile, err = os.OpenFile("build", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
			bwrite := func(format string, args ...interface{}) {
				if err != nil {
					return
				}
				_, err = fmt.Fprintf(buildFile, format, args...)
			}
			bwrite("# Build script generated by gb: http://go-gb.googlecode.com\n")
			bwrite("# gb provides configuration-free building and distributing\n")
			bwrite("\n")
			bwrite("echo \"Build script generated by gb: http://go-gb.googlecode.com\" \n")

			gm := make(map[string]bool)
			for _, pkg := range ListedPkgs {
				pkg.CollectGoInstall(gm)
			}
			bwrite("if [ \"$1\" = \"goinstall\" ]; then\n")
			bwrite("echo Running goinstall \\\n")
			for gp := range gm {
				if _, ok := Packages[gp]; ok {
					continue
				}
				gp = strings.Trim(gp, "\"")
				bwrite("&& echo \"goinstall %s\" \\\n", gp)
				bwrite("&& goinstall %s \\\n", gp)
			}
			bwrite("\n")
			bwrite("else\n")
			bwrite("echo Building \\\n")
			for _, pkg := range ListedPkgs {
				pkg.AddToBuild(buildFile)
			}
			bwrite("\n")
			bwrite("fi\n")
			bwrite("\n# The makefiles above are invoked in topological dependence order\n")

			if err != nil {
				return
			}

			err = buildFile.Close()
			if err != nil {
				return
			}
		}
		for _, pkg := range ListedPkgs {
			err = pkg.GenerateMakefile()
			if err != nil {
				return
			}
		}
	}

	return
}

func TryDistribution() (err error) {
	if Distribution {
		err = errors.New("the '--dist' feature has been removed - use your version control's archive utility")
	}
	return
}

func TryClean() {
	if Clean && ListedTargets == 0 {
		fmt.Println("Removing " + GetBuildDirPkg())
		os.RemoveAll(GetBuildDirPkg())
		fmt.Println("Removing " + GetBuildDirCmd())
		os.RemoveAll(GetBuildDirCmd())
		PackagesCleaned++
	}
	if Clean && len(ListedDirs) == 1 {
		var dir string
		for d := range ListedDirs {
			dir = d
		}
		base := filepath.Base(dir)
		if base == "testdata" {
			testObj := filepath.Join(dir, "_obj")
			testBin := filepath.Join(dir, "_bin")
			fmt.Println("Removing " + testObj)
			os.RemoveAll(testObj)
			fmt.Println("Removing " + testBin)
			os.RemoveAll(testBin)
			PackagesCleaned++
		}
	}

	if Clean {
		for _, pkg := range ListedPkgs {
			pkg.Clean()
		}
	}
}

func TryBuild() {

	if Build {
		if Concurrent {
			for _, pkg := range ListedPkgs {
				pkg.CheckStatus()
				go pkg.Build()
			}
		}
		for _, pkg := range ListedPkgs {
			pkg.CheckStatus()
			err := pkg.Build()
			if err != nil {
				return
			}
		}
	}
}

func TryTest() (err error) {
	if Test {
		for _, pkg := range ListedPkgs {
			if len(pkg.TestSources) != 0 {
				err = pkg.Test()
				if err != nil {
					return
				}
			}
		}
	}
	return
}

func TryInstall() {
	if Install {
		brokenMsg := []string{}
		for _, pkg := range ListedPkgs {
			err := pkg.Install()
			if err != nil {
				brokenMsg = append(brokenMsg, fmt.Sprintf("(in %s) could not install \"%s\"", pkg.Dir, pkg.Target))
			}
		}

		if len(brokenMsg) != 0 {
			for _, msg := range brokenMsg {
				fmt.Printf("%s\n", msg)
			}
		}
	}
}

func RunGB() (err error) {
	Build = Build || (!Clean && !Scan) || (Makefiles && !Clean) || Install || Test

	Build = Build && HardArgs == 0

	DoPkgs, DoCmds = DoPkgs || (!DoPkgs && !DoCmds), DoCmds || (!DoPkgs && !DoCmds)

	ListedDirs = make(map[string]bool)
	ValidatedDirs = make(map[string]bool)

	args := os.Args[1:len(os.Args)]

	err = ScanDirectory(".", ".", "")
	if err != nil {
		return
	}
	if BuildGOROOT {
		fmt.Printf("Scanning %s...", filepath.Join("GOROOT", "src"))
		ScanDirectory("", filepath.Join(GOROOT, "src"), "")
		fmt.Printf("done\n")
		for _, gp := range GOPATHS {
			fmt.Printf("Scanning %s...", filepath.Join(gp, "src"))
			ScanDirectory("", filepath.Join(gp, "src"), "")
			fmt.Printf("done\n")
		}
	}

	for _, arg := range args {
		if arg[0] != '-' {
			carg := filepath.Clean(arg)
			rel := GetRelative(CWD, carg, OSWD)
			ListedDirs[rel] = true
			ListedTargets++
		}
	}

	if ListedTargets == 0 {
		rel := GetRelative(CWD, OSWD, OSWD)
		if rel != "." {
			ListedDirs[GetRelative(CWD, OSWD, OSWD)] = true
			ListedTargets++
		}
	}

	ListedPkgs = []*Package{}
	for _, pkg := range Packages {
		if !RunningInGOROOT && pkg.IsInGOROOT {
			continue
		}
		if RunningInGOPATH == "" && pkg.IsInGOPATH != "" {
			continue
		}
		if IsListed(pkg.Dir) {
			ListedPkgs = append(ListedPkgs, pkg)
			ValidateDir(pkg.Dir)
		}
	}

	for lt := range ListedDirs {
		if !ValidatedDirs[lt] {
			err = errors.New(fmt.Sprintf("Listed directory %q doesn't correspond to a known package", lt))
			return
		}
	}

	if len(ListedPkgs) == 0 {
		err = errors.New("No targets found in " + CWD)
		return
	}

	for _, pkg := range Packages {
		pkg.Stat()
	}

	for _, pkg := range Packages {
		pkg.ResolveDeps()
	}

	for _, pkg := range Packages {
		cycle := pkg.DetectCycles()
		if cycle != nil {
			var targets []string
			for _, cp := range cycle {
				targets = append(targets, cp.Target)
			}
			err = errors.New(fmt.Sprintf("Cycle detected: %v", targets))
			return
		}
	}

	for _, pkg := range Packages {
		pkg.CheckStatus()
	}

	TryScan()

	if err = TryGoFix(); err != nil {
		return
	}

	if err = TryGoFMT(); err != nil {
		return
	}

	if err = TryGenMake(); err != nil {
		return
	}

	if err = TryDistribution(); err != nil {
		return
	}

	TryClean()

	TryBuild()

	if err = TryTest(); err != nil {
		return
	}

	TryInstall()

	if Build {
		if PackagesBuilt > 1 {
			fmt.Printf("Built %d targets\n", PackagesBuilt)
		} else if PackagesBuilt == 1 {
			fmt.Printf("Built 1 target\n")
		}
		if PackagesInstalled > 1 {
			fmt.Printf("Installed %d targets\n", PackagesInstalled)
		} else if PackagesInstalled == 1 {
			fmt.Println("Installed 1 target")
		}
		if Build && PackagesBuilt == 0 && PackagesInstalled == 0 && BrokenPackages == 0 {
			fmt.Println("Up to date")
		}
		if BrokenPackages > 1 {
			fmt.Printf("%d broken targets\n", BrokenPackages)
		} else if BrokenPackages == 1 {
			fmt.Println("1 broken target")
		}
		if len(BrokenMsg) != 0 {
			for _, msg := range BrokenMsg {
				fmt.Printf("%s\n", msg)
			}
		}
	}
	if Clean {
		if PackagesCleaned == 0 {
			fmt.Printf("No mess to clean\n")
		}
	}

	return
}

func CheckFlags() bool {
	for i, arg := range os.Args[1:] {
		if arg == "--testargs" {
			TestArgs = append(TestArgs, os.Args[i+2:]...)
			os.Args = os.Args[:i+2]
			if !Test {
				ErrLog.Printf("Must be in test mode (-t) to use --testargs")
				return false
			}
			break
		}
		if strings.HasPrefix(arg, "-test.") {
			TestArgs = append(TestArgs, arg)
			continue
		}
		if strings.HasPrefix(arg, "--") {
			switch arg {
			case "--gofmt":
				GoFMT = true
			case "--gofix":
				GoFix = true
			case "--dist":
				Distribution = true
			case "--makefiles":
				GenMake = true
			case "--workspace":
				Workspace = true
			default:
				Usage()
				return false
			}
			HardArgs++
		} else if strings.HasPrefix(arg, "-") {
			for _, flag := range arg[1:] {
				switch flag {
				case 'i':
					Install = true
					BuildArgs++
				case 'c':
					Clean = true
				case 'b':
					Build = true
					BuildArgs++
				case 's':
					Scan = true
				case 'S':
					Scan = true
					ScanList = true
				case 'L':
					Scan = true
					ScanListFiles = true
				case 't':
					Test = true
					BuildArgs++
				case 'e':
					Exclusive = true
				case 'v':
					Verbose = true
				case 'm':
					Makefiles = true
				case 'f':
					Force = true
				case 'g':
					GoInstall = true
					BuildArgs++
				case 'G':
					GoInstall = true
					GoInstallUpdate = true
					BuildArgs++
				case 'p':
					Concurrent = true
				case 'P':
					DoPkgs = true
				case 'C':
					DoCmds = true
				case 'R':
					BuildGOROOT = true
				case 'N':
					Clean = true
					Nuke = true
				default:
					Usage()
					return false
				}
			}
		}
	}

	if HardArgs > 0 && BuildArgs > 0 {
		ErrLog.Printf("Cannot have -- style arguments and build at the same time.\n")
		return false
	}

	return true
}

func main() {
	if err := LoadCWD(); err != nil {
		ErrLog.Printf("%v\n", err)
		return
	}

	if !LoadEnvs() {
		return
	}

	if !CheckFlags() {
		return
	}

	err := FindExternals()
	if err != nil {
		return
	}

	GCArgs = []string{}
	GLArgs = []string{}

	if !Install {
		IncludeDir = GetBuildDirPkg()
		GCArgs = append(GCArgs, []string{"-I", IncludeDir}...)
		GLArgs = append(GLArgs, []string{"-L", IncludeDir}...)
	}

	err = RunGB()
	if err != nil {
		ErrLog.Printf("%v\n", err)
		ReturnFailCode = true
	}

	if len(BrokenMsg) > 0 {
		ReturnFailCode = true
	}

	if ReturnFailCode {
		os.Exit(1)
	}
}
