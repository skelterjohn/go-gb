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
	"fmt"
	"log"
	"os"
	"path"
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

 math has both pure go and asm versions of many functions, and which is used depends
 on the architexture

 go/build has a source-generation step that uses make variables

 os has source generation

 syscall has crazy pure go/asm versions and unused source files

 crypto/tls has a file root_stub.go which is excluded
*/
var ForceMakePkgs = map[string]bool{
	"math":       true,
	"go/build":   true,
	"os":         true,
	"hash/crc32": true,
	"syscall":    true,
	"runtime":    true,
	"crypto/tls": true,
	"godoc":      true,
}

var OSFiltersMust = map[string]string{
	"wingui": "windows",
}

func ScanDirectory(base, dir string) (err2 os.Error) {
	_, basedir := path.Split(dir)
	if basedir == "_obj" ||
		basedir == "_test" ||
		basedir == "_cgo" ||
		basedir == "bin" ||
		(basedir != "." && strings.HasPrefix(basedir, ".")) {
		return
	}

	var err os.Error

	cfg := ReadConfig(dir)


	if Workspace {
		absdir := GetAbs(dir, CWD)
		relworkspace := GetRelative(absdir, CWD, CWD)

		cfg["workspace"] = relworkspace
		cfg.Write(absdir)
	}



	var pkg *Package

	if ignore, ok := cfg.Ignore(); !(ignore && ok) {
		pkg, err = NewPackage(base, dir, cfg)
		if err == nil {
			/*
			if Workspace {
				absdir := GetAbs(dir, CWD)
				relworkspace := GetRelative(absdir, CWD, CWD)
				var wfile *os.File
				wfile, err = os.Create(path.Join(absdir, "workspace.gb"))
				wfile.WriteString(relworkspace + "\n")
				wfile.Close()		
			}
			*/

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

	/*
	if pkg == nil {
		return
	}
	*/

	if pkg != nil && pkg.Target == "." {
		err = os.NewError("Package has no name specified. Either create 'target.gb' or run gb from above.")
	}

	if ignoreAll, ok := cfg.IgnoreAll(); !(ignoreAll && ok) {
		subdirs := GetSubDirs(dir)
		for _, subdir := range subdirs {
			ScanDirectory(path.Join(base, subdir), path.Join(dir, subdir))
		}
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

func TryGoFMT() (err os.Error) {
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

func TryGoFix() (err os.Error) {
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

func TryGenMake() (err os.Error) {
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

func TryDistribution() (err os.Error) {
	if Distribution {
		err = os.NewError("the '--dist' feature has been removed - use your version control's archive utility")
	}
	return
}

func TryClean() {
	if Clean && ListedTargets == 0 {
		fmt.Println("Removing " + GetBuildDirPkg())
		os.RemoveAll(GetBuildDirPkg())
		fmt.Println("Removing " + GetBuildDirCmd())
		os.RemoveAll(GetBuildDirCmd())
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

func TryTest() (err os.Error) {
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

func RunGB() (err os.Error) {
	Build = Build || (!Clean && !Scan) || (Makefiles && !Clean) || Install || Test

	Build = Build && HardArgs == 0

	DoPkgs, DoCmds = DoPkgs || (!DoPkgs && !DoCmds), DoCmds || (!DoPkgs && !DoCmds)

	ListedDirs = make(map[string]bool)
	ValidatedDirs = make(map[string]bool)

	args := os.Args[1:len(os.Args)]

	err = ScanDirectory(".", ".")
	if err != nil {
		return
	}
	if BuildGOROOT {
		fmt.Printf("Scanning %s...", path.Join("GOROOT", "src"))
		ScanDirectory("", path.Join(GOROOT, "src"))
		fmt.Printf("done\n")
		for _, gp := range GOPATHS {
			fmt.Printf("Scanning %s...", path.Join(gp, "src"))
			ScanDirectory("", path.Join(gp, "src"))
			fmt.Printf("done\n")
		}
	}

	for _, arg := range args {
		if arg[0] != '-' {
			carg := path.Clean(arg)
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
			err = os.NewError(fmt.Sprintf("Listed directory %q doesn't correspond to a known package", lt))
			return
		}
	}

	if len(ListedPkgs) == 0 {
		err = os.NewError("No targets found in " + CWD)
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
			err = os.NewError(fmt.Sprintf("Cycle detected: %v", targets))
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
		if PackagesBuilt == 0 {
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
