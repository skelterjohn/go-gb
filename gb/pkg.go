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
	"bufio"
	"strings"
	"path"
)

type Package struct {
	Dir, Base         string
	
	Name, Target      string
	
	IsCmd       bool
	Active      bool
	
	ResultPath, InstallPath string

	IsCGo bool
	
	//these prevent multipath issues for tree following
	built, cleaned, addedToBuild, gofmted, scanned bool

	NeedsBuild, NeedsInstall, NeedsGoInstall bool

	Sources    []string
	CGoSources []string
	CSrcs      []string
	AsmSrcs    []string

	Objects []string

	PkgSrc  map[string][]string
	TestSrc map[string][]string
	
	SrcDeps map[string][]string
	Deps        []string
	DepPkgs     []*Package

	TestSources []string
	TestDeps    []string
	TestFuncs map[string][]string
	TestDepPkgs []*Package

	HasMakefile bool
	IsInGOROOT  bool


	SourceTime, BinTime, InstTime, GOROOTPkgTime int64

	block chan bool
}

func ReadPackage(base, dir string) (this *Package, err os.Error) {
	//println("ReadPackage(",base,dir,")")
	finfo, err := os.Stat(dir)
	if err != nil || !finfo.IsDirectory() {
		err = os.NewError("not a directory")
		return
	}
	this = new(Package)
	this.block = make(chan bool, 1)
	this.Dir = path.Clean(dir)
	this.PkgSrc = make(map[string][]string)
	this.TestSrc = make(map[string][]string)
	this.TestFuncs = make(map[string][]string)

	//global, _ := GetAbsolutePath(dir)
	//if strings.HasPrefix(global, GOROOT) {
	//println("rp: ", GOROOT, dir)
	cwd, _ := os.Getwd()
	if rel := GetRelative(GOROOT, dir, cwd); !strings.HasPrefix(rel, "..") {
		this.IsInGOROOT = true
	}

	err = this.ScanForSource()
	if err != nil {
		return
	}
	err = this.GetSourceDeps()
	if err != nil {
		return
	}

	this.Base = base
	this.DepPkgs = make([]*Package, 0)

	if strings.HasPrefix(this.Dir, "./") {
		this.Dir = this.Dir[2:len(this.Dir)]
	}

	if _, err2 := os.Stat(path.Join(this.Dir, "/Makefile")); err2 == nil {
		this.HasMakefile = true
	}
	if _, err2 := os.Stat(path.Join(this.Dir, "/makefile")); err2 == nil {
		this.HasMakefile = true
	}

	for _, src := range this.Sources {
		var t int64
		t, err = StatTime(path.Join(this.Dir, src))
		if err != nil {
			return
		}
		if t > this.SourceTime {
			this.SourceTime = t
		}
	}

	if err != nil {
		return
	}
	this.IsCmd = this.Name == "main"
	this.Objects = append(this.Objects, path.Join(this.Dir, GetIBName()))
	err = this.GetTarget()

	this.Active = (DoCmds && this.IsCmd) || (DoPkgs && !this.IsCmd)

	return
}

func (this *Package) ScanForSource() (err os.Error) {
	errch := make(chan os.Error)
	path.Walk(this.Dir, this, errch)

	if len(this.Sources)+len(this.TestSources) == 0 {
		err = os.NewError("No source files in " + this.Dir)
	}

	this.IsCGo = this.IsCGo || len(this.CSrcs) /*+len(this.AsmSrcs)*/ > 0

	return
}
func (this *Package) VisitDir(dpath string, f *os.FileInfo) bool {
	return dpath == this.Dir || strings.HasPrefix(dpath, path.Join(this.Dir, "src"))
}
func (this *Package) VisitFile(fpath string, f *os.FileInfo) {
	if !FilterFlag(fpath) {
		return
	}
	if strings.HasSuffix(fpath, "_testmain.go") {
		return
	}
	rootl := len(this.Dir) + 1
	if this.Dir != "." {
		fpath = fpath[rootl:len(fpath)]
	}
	if strings.HasSuffix(fpath, ".s") {
		this.AsmSrcs = append(this.AsmSrcs, fpath)

		this.Objects = append(this.Objects, fpath[:len(fpath)-2]+GetObjSuffix())
	}
	if strings.HasSuffix(fpath, ".go") {
		if strings.HasSuffix(fpath, "_test.go") {
			this.TestSources = append(this.TestSources, fpath)
		} else if strings.HasPrefix(fpath, "cgo_") {
			this.CGoSources = append(this.CGoSources, fpath)
		} else {
			this.Sources = append(this.Sources, fpath)
		}
	}
	if strings.HasSuffix(fpath, ".c") {
		this.CSrcs = append(this.CSrcs, fpath)
	}
}

func (this *Package) GetSourceDeps() (err os.Error) {
	this.SrcDeps = make(map[string][]string)
	for _, src := range this.Sources {
		var fpkg, ftarget string
		var fdeps []string
		fpkg, ftarget, fdeps, _, err = GetDeps(path.Join(this.Dir, src))

		if err != nil {
			BrokenMsg = append(BrokenMsg, fmt.Sprintf("(in %s) %s", this.Dir, err.String()))
			return
		}

		this.PkgSrc[fpkg] = append(this.PkgSrc[fpkg], src)

		this.SrcDeps[src] = fdeps

		if ftarget != "" {
			this.Target = ftarget
		}
		if fpkg != "documentation" {
			if fpkg != "main" || this.Name == "" {
				this.Name = fpkg
			}
		}
		//this.Deps = append(this.Deps, fdeps...)
	}

	for _, buildSrc := range this.PkgSrc[this.Name] {
		this.Deps = append(this.Deps, this.SrcDeps[buildSrc]...)
	}

	this.Deps = RemoveDups(this.Deps)
	if Test {
		for _, src := range this.TestSources {
			var fpkg, ftarget string
			var fdeps, ffuncs []string
			fpkg, ftarget, fdeps, ffuncs, err = GetDeps(path.Join(this.Dir, src))
			//this.SrcDeps[src] = fdeps
			this.TestSrc[fpkg] = append(this.TestSrc[fpkg], src)
			if err != nil {
				BrokenMsg = append(BrokenMsg, fmt.Sprintf("(in %s) %s", this.Dir, err.String()))
				break
			}
			if ftarget != "" {
				this.Target = ftarget
			}
			//this.Name = fpkg
			this.TestDeps = append(this.TestDeps, fdeps...)
			//this.Funcs = append(this.Funcs, ffuncs...)
			this.TestFuncs[fpkg] = append(this.TestFuncs[fpkg], ffuncs...)
		}
		this.TestDeps = RemoveDups(this.TestDeps)
	}
	return
}

func (this *Package) GetTarget() (err os.Error) {
	if !this.IsCmd && this.IsInGOROOT {
		//always the relative path
		//println("grp:", path.Join(GOROOT, "src", "pkg"), this.Dir)
		cwd, _ := os.Getwd()
		this.Target = GetRelative(path.Join(GOROOT, "src", "pkg"), this.Dir, cwd)
		if strings.HasPrefix(this.Target, "..") {
			err = os.NewError(fmt.Sprintf("(in %s) GOROOT pkg is not in $GOROOT/src/pkg", this.Dir))
			return
		}

		//fmt.Printf("found goroot relative path for %s = %s\n", this.Dir, this.Target)
	} else {
		if this.Target == "" {
			this.Target = this.Base

			if this.IsCmd {
				this.Target = path.Base(this.Dir)
				if this.Target == "." {
					this.Target = "main"
				}
			} else {
				if this.Target == "." {
					this.Target = "localpkg"
				}
			}
		} else {
			this.Base = this.Target
		}

		tpath := path.Join(this.Dir, "/target.gb")
		fin, err2 := os.Open(tpath, os.O_RDONLY, 0)
		if err2 == nil {
			bfrd := bufio.NewReader(fin)
			this.Target, err = bfrd.ReadString('\n')
			this.Target = strings.TrimSpace(this.Target)
			this.Base = this.Target
			if this.Target == "-" {
				err = os.NewError("directory opts-out")
				return
			}
		}
	}

	this.Base = path.Clean(this.Base)
	this.Target = path.Clean(this.Target)

	err = nil
	if this.IsCmd {
		if GOOS == "windows" {
			this.Target += ".exe"
		}
		this.InstallPath = path.Join(GetInstallDirCmd(), this.Target)
		this.ResultPath = path.Join(GetBuildDirCmd(), this.Target)
	} else {

		this.InstallPath = path.Join(GetInstallDirPkg(), this.Target+".a")
		this.ResultPath = path.Join(GetBuildDirPkg(), this.Target+".a")
	}

	if this.IsInGOROOT {
		this.ResultPath = this.InstallPath
	}

	this.Stat()

	return
}

func (this *Package) PrintScan() {
	if this.scanned {
		return
	}
	this.scanned = true

	for _, pkg := range this.DepPkgs {
		pkg.PrintScan()
	}

	//build, install := this.Touched()
	bis := ""
	if !this.NeedsBuild {
		bis = " (up to date)"
	}
	if !this.NeedsInstall {
		bis = " (installed)"
	}
	var label string

	if this.IsCmd {
		label = "cmd"
	} else {
		label = "pkg"
	}
	if this.IsCGo {
		label = "cgo"
	}
	if this.IsInGOROOT {
		label = "goroot " + label
	}

	displayDir := this.Dir
	if this.IsInGOROOT {
		displayDir = strings.Replace(displayDir, GOROOT, "$GOROOT", 1)
	}
	var prefix string
	if !this.IsInGOROOT {
		prefix = fmt.Sprintf("in %s: ", displayDir)
	}
	fmt.Printf("%s%s \"%s\"%s\n", prefix, label, this.Target, bis)
	if ScanList {
		fmt.Printf(" %s Deps: %v\n", this.Target, this.Deps)
		fmt.Printf(" %s TestDeps: %v\n", this.Target, this.TestDeps)
	}
}

func (this *Package) Stat() {
	this.BinTime, _ = StatTime(this.ResultPath)
	this.InstTime, _ = StatTime(this.InstallPath)
	/*
		resInfo, err := os.Stat(this.ResultPath)
		if resInfo != nil && err == nil {
			this.BinTime = resInfo.Mtime_ns
		} else {
			this.BinTime = 0
		}
		resInfo, err = os.Stat(this.InstallPath)
		if resInfo != nil && err == nil {
			this.InstTime = resInfo.Mtime_ns
		} else {
			this.InstTime = 0
		}
	*/
}

func (this *Package) CheckStatus() {
	b, i := this.Touched()
	this.NeedsBuild = b || this.NeedsBuild
	this.NeedsInstall = i || this.NeedsInstall
}

func (this *Package) ResolveDeps() (err os.Error) {
	CheckDeps := func(deps []string, test bool) (err os.Error) {
		for _, dep := range deps {
			if dep == "\"C\"" {
				this.IsCGo = true
				continue
			}
			if pkg, ok := Packages[dep]; ok {
				if test {
					this.TestDepPkgs = append(this.TestDepPkgs, pkg)
				} else {
					this.DepPkgs = append(this.DepPkgs, pkg)
				}
			} else {
				exists, when := PkgExistsInGOROOT(dep)
				if exists {
					if this.GOROOTPkgTime < when {
						this.GOROOTPkgTime = when
					}
				}
				if !IsGoInstallable(dep) {
					if !exists {
						//fmt.Printf("in %s: can't resolve pkg %s (maybe you aren't in the root?)\n", this.Dir, dep)
						err = os.NewError("unresolved packages")
					}
				} else {
					if GoInstallUpdate {
						this.NeedsBuild = true
					}
					if !exists {
						if !GoInstall {
							//fmt.Printf("in %s: can't resolve pkg %s (try using -g)\n", this.Dir, dep)
							err = os.NewError("unresolved packages")
						} else {
							this.NeedsGoInstall = true
							this.NeedsBuild = true
						}
					}

				}
			}
		}
		return
	}
	err = CheckDeps(this.Deps, false)
	if err != nil {
		return
	}
	err = CheckDeps(this.TestDeps, true)
	return
}

func (this *Package) Touched() (build, install bool) {
	var inTime int64

	build = this.NeedsBuild
	install = this.NeedsInstall

	for _, pkg := range this.DepPkgs {
		db, di := pkg.Touched()
		if db {
			build = true
		}
		if di {
			install = true
		}
		if pkg.BinTime > inTime {
			inTime = pkg.BinTime
		}
	}
	if this.GOROOTPkgTime > inTime {
		inTime = this.GOROOTPkgTime
	}

	if this.SourceTime > inTime {
		inTime = this.SourceTime
	}
	if inTime > this.BinTime {
		build = true
	}
	if this.InstTime < this.BinTime || this.InstTime < inTime {
		install = true
	}

	if build {
		install = true
	}

	return
}

func (this *Package) Build() (err os.Error) {
	this.block <- true
	defer func() {
		<-this.block
	}()

	if !this.NeedsBuild {
		return
	}
	if this.built {
		return
	}
	this.built = true
	
	if !this.HasMakefile && this.IsCGo {
		fmt.Printf("(in %s) this is a cgo project; please create a makefile", this.Dir)
		return
	}

	if Exclusive && !ListedDirs[this.Dir] {
		return
	}

	inTime := this.GOROOTPkgTime

	if Concurrent {
		for _, pkg := range this.DepPkgs {
			go pkg.Build()
		}
	}

	for _, pkg := range this.DepPkgs {

		err = pkg.Build()
		if err != nil {
			return
		}
		if pkg.BinTime > inTime {
			inTime = pkg.BinTime
		}
	}
	if GoInstall {
		for _, dep := range this.Deps {
			if _, ok := Packages[dep]; !ok {
				goinstTime := GoInstallPkg(dep)
				if goinstTime > inTime {
					inTime = goinstTime
				}
			}
		}
	}

	if !this.Active {
		return
	}

	if this.SourceTime > inTime {
		inTime = this.SourceTime
	}

	if inTime > this.BinTime {
		which := "cmd"
		if this.Name != "main" {
			which = "pkg"
		}
		fmt.Printf("(in %s) building %s \"%s\"\n", this.Dir, which, this.Target)

		if Makefiles && this.HasMakefile {
			err = MakeBuild(this)
		} else if this.IsCGo {
			err = BuildCgoPackage(this)
		} else {
			err = BuildPackage(this)
		}

		if err == nil {
			PackagesBuilt++
		} else {
			BrokenPackages++
			BrokenMsg = append(BrokenMsg, fmt.Sprintf("(in %s) could not build \"%s\"", this.Dir, this.Target))
		}

	}
	if err != nil {
		this.CleanFiles()
	}

	if this.IsInGOROOT && this.HasMakefile {
		err = this.Install()
	}

	this.NeedsBuild = false
	this.Stat()

	return
}
func (this *Package) Test() (err os.Error) {
	for _, pkg := range this.TestDepPkgs {
		err = pkg.Build()
		if err != nil {
			return
		}
	}
	if GoInstall {
		for _, dep := range this.TestDeps {
			if _, ok := Packages[dep]; !ok {
				GoInstallPkg(dep)
			}
		}
	}

	if (Makefiles && this.HasMakefile) || this.IsCGo {
		err = MakeTest(this)
		return
	}

	testdir := path.Join(this.Dir, "_test")
	if Verbose {
		fmt.Printf(" Removing %s\n", testdir)
	}
	err = os.RemoveAll(testdir)

	fmt.Printf("(in %s) testing \"%s\"\n", this.Dir, this.Target)

	var pkgtests, pkgbenchmarks map[string][]string
	pkgtests = make(map[string][]string)
	pkgbenchmarks = make(map[string][]string)

	for name, funcs := range this.TestFuncs {
		for _, f := range funcs {
			if strings.HasPrefix(f, "Test") {
				pkgtests[name] = append(pkgtests[name], f)
			}
			if strings.HasPrefix(f, "Benchmark") {
				pkgbenchmarks[name] = append(pkgbenchmarks[name], f)
			}
		}
	}

	testsrc := path.Join(this.Dir, "_test", "_testmain.go")
	dstDir, _ := path.Split(testsrc)
	os.MkdirAll(dstDir, 0755)
	file, err := os.Open(testsrc, os.O_CREATE|os.O_RDWR, 0644)

	if err != nil {
		return
	}

	fmt.Fprintf(file, "package main\n\n")

	//fmt.Fprintf(file, "import \"%s\"\n", this.Target)
	for name, _ := range this.TestSrc {
		if name == "main" {
			fmt.Fprintf(file, "import __main__ \"_test/%s\"\n", name)
		} else {
			fmt.Fprintf(file, "import \"_test/%s\"\n", name)
		}
	}
	fmt.Fprintf(file, "import \"testing\"\n")
	fmt.Fprintf(file, "import __regexp__ \"regexp\"\n\n")

	fmt.Fprintf(file, "var tests = []testing.InternalTest{\n")
	for name, tests := range pkgtests {
		for _, test := range tests {
			callName := name
			if name == "main" {
				callName = "__main__"
			}
			fmt.Fprintf(file, "\t{\"%s.%s\", %s.%s},\n", name, test, callName, test)
		}
	}
	fmt.Fprintf(file, "}\n")

	fmt.Fprintf(file, "var benchmarks = []testing.InternalBenchmark{\n")
	for name, benchmarks := range pkgbenchmarks {
		for _, benchmark := range benchmarks {
			fmt.Fprintf(file, "\t{\"%s.%s\", %s.%s},\n", name, benchmark, name, benchmark)
		}
	}
	fmt.Fprintf(file, "}\n\n")

	fmt.Fprintf(file, "func main() {\n")
	fmt.Fprintf(file, "\ttesting.Main(__regexp__.MatchString, tests)\n")
	fmt.Fprintf(file, "\ttesting.RunBenchmarks(__regexp__.MatchString, benchmarks)\n")
	fmt.Fprintf(file, "}\n")

	file.Close()

	err = BuildTest(this)

	this.Stat()

	return
}
/*
package main

import "go-glue.googlecode.com/hg/rlglue"
import "testing"
import __regexp__ "regexp"

var tests = []testing.InternalTest{
	{"rlglue.TestTaskSpec", rlglue.TestTaskSpec},
}
var benchmarks = []testing.InternalBenchmark{}

func main() {
	testing.Main(__regexp__.MatchString, tests)
	testing.RunBenchmarks(__regexp__.MatchString, benchmarks)
}

*/

func (this *Package) CleanFiles() (err os.Error) {
	defer func() {
		this.Stat()
		this.NeedsBuild = true
		this.NeedsInstall = true
	}()

	if (Makefiles && this.HasMakefile) || this.IsCGo {
		MakeClean(this)
		PackagesBuilt++
		return
	}

	if Nuke {
		if _, err2 := os.Stat(this.InstallPath); err2 == nil {
			reallyDoIt := true
			if !Force {
				fmt.Printf("Really nuke installed binary '%s'? (y/n) ", this.InstallPath)
				var answer string
				fmt.Scanf("%s", &answer)
				reallyDoIt = answer == "y" || answer == "Y"
			}
			if reallyDoIt {
				if Verbose {
					fmt.Printf(" Removing %s\n", this.InstallPath)
				}
				err = os.Remove(this.InstallPath)
			}
		}
	}

	ib := false
	res := false
	test := false
	for _, obj := range this.Objects {
		if _, err2 := os.Stat(obj); err2 == nil {
			ib = true
		}
	}
	if _, err2 := os.Stat(this.ResultPath); err2 == nil {
		res = true
	}
	testdir := path.Join(this.Dir, "_test")
	if _, err2 := os.Stat(testdir); err2 == nil {
		test = true
	}
	if !ib && !res && !test {
		return
	}
	fmt.Printf("Cleaning %s\n", this.Dir)
	for _, obj := range this.Objects {
		if Verbose {
			fmt.Printf(" Removing %s\n", obj)
		}
		err = os.Remove(obj)
	}
	if Verbose {
		fmt.Printf(" Removing %s\n", this.ResultPath)
	}
	err = os.Remove(this.ResultPath)
	if Verbose {
		fmt.Printf(" Removing %s\n", testdir)
	}
	err = os.RemoveAll(testdir)

	return
}

func (this *Package) Clean() (err os.Error) {
	if this.cleaned {
		return
	}
	this.cleaned = true
	if Exclusive && !ListedDirs[this.Dir] {
		return
	}

	for _, pkg := range this.DepPkgs {
		pkg.Clean()
	}

	if !this.Active {
		return
	}

	err = this.CleanFiles()

	PackagesBuilt++

	return
}
func (this *Package) Install() (err os.Error) {
	if !this.NeedsInstall {
		return
	}
	if Exclusive && !ListedDirs[this.Dir] {
		return
	}

	for _, pkg := range this.DepPkgs {
		pkg.Install()
	}

	if !this.Active {
		return
	}

	if !(Makefiles && this.HasMakefile) && this.InstTime < this.BinTime && !this.IsInGOROOT {
		err = InstallPackage(this)

		this.Stat()

		PackagesInstalled++
	}
	return
}

func (this *Package) CollectDistributionFiles(ch chan string) (err os.Error) {
	if Exclusive && !ListedDirs[this.Dir] {
		return
	}
	var f string
	f = path.Join(this.Dir, "Makefile")
	if _, err2 := os.Stat(f); err2 == nil {
		ch <- f
	}
	f = path.Join(this.Dir, "target.gb")
	if _, err2 := os.Stat(f); err2 == nil {
		ch <- f
	}
	f = path.Join(this.Dir, "README")
	if _, err2 := os.Stat(f); err2 == nil {
		ch <- f
	}
	for _, src := range this.Sources {
		ch <- path.Join(this.Dir, src)
	}
	for _, src := range this.CSrcs {
		ch <- path.Join(this.Dir, src)
	}
	for _, src := range this.CGoSources {
		ch <- path.Join(this.Dir, src)
	}
	for _, src := range this.TestSources {
		ch <- path.Join(this.Dir, src)
	}

	for _, pkg := range this.DepPkgs {
		err = pkg.CollectDistributionFiles(ch)
		if err != nil {
			return
		}
	}

	for _, pkg := range this.TestDepPkgs {
		err = pkg.CollectDistributionFiles(ch)
		if err != nil {
			return
		}
	}

	return
}

/*
include $(GOROOT)/src/Make.inc

TARG=gb
GOFILES=\
	gb.go\
	deps.go\
	build.go\
	make.go\
	pkg.go\
	goinstall.go\

include $(GOROOT)/src/Make.cmd

*/
func (this *Package) GenerateMakefile() (err os.Error) {
	if !this.Active {
		return
	}

	if this.IsCGo {
		fmt.Printf("(in %s) this is a cgo project; skipping makefile generation\n", this.Dir)
		return
	}

	mpath := path.Join(this.Dir, "Makefile")

	_, ferr := os.Stat(mpath)
	if ferr == nil {
		if !Force {
			fmt.Printf("'%s' exists; overwrite? (y/n) ", mpath)
			var answer string
			fmt.Scanf("%s", &answer)
			if answer != "y" && answer != "Y" {
				err = nil
				return
			}
		}
		os.Remove(mpath)
	}

	which := "pkg"
	if this.IsCmd {
		which = "cmd"
	}
	fmt.Printf("(in %s) generating makefile for %s \"%s\"\n", this.Dir, which, this.Target)

	var file *os.File
	file, err = os.Open(mpath, os.O_CREATE|os.O_RDWR, 0644)

	if err != nil {
		return
	}

	_, err = fmt.Fprintf(file, "# Makefile generated by gb: http://go-gb.googlecode.com\n")
	_, err = fmt.Fprintf(file, "# gb provides configuration-free building and distributing\n")
	_, err = fmt.Fprintf(file, "\n")
	_, err = fmt.Fprintf(file, "include $(GOROOT)/src/Make.inc\n")
	_, err = fmt.Fprintf(file, "\n")
	makeTarget := this.Target
	if GOOS == "windows" && strings.HasSuffix(makeTarget, ".exe") {
		makeTarget = makeTarget[0 : len(makeTarget)-len(".exe")]
	}
	_, err = fmt.Fprintf(file, "TARG=%s\n", makeTarget)
	_, err = fmt.Fprintf(file, "GOFILES=\\\n")
	for _, src := range this.PkgSrc[this.Name] {
		_, err = fmt.Fprintf(file, "\t%s\\\n", src)
	}
	_, err = fmt.Fprintf(file, "\n")

	reverseDots := ReverseDir(this.Dir)

	_, err = fmt.Fprintf(file, "# gb: this is the local install\n")
	_, err = fmt.Fprintf(file, "GBROOT=%s\n", reverseDots)
	_, err = fmt.Fprintf(file, "\n")
	relObj := path.Join("$(GBROOT)", GetBuildDirPkg())
	_, err = fmt.Fprintf(file, "# gb: compile/link against local install\n")
	_, err = fmt.Fprintf(file, "GC+= -I %s\n", relObj)
	_, err = fmt.Fprintf(file, "LD+= -L %s\n", relObj)
	_, err = fmt.Fprintf(file, "\n")

	if this.IsCmd {
		_, err = fmt.Fprintf(file, "# gb: default target is in GBROOT this way\n")
		_, err = fmt.Fprintf(file, "command:\n")
		_, err = fmt.Fprintf(file, "\n")
		_, err = fmt.Fprintf(file, "include $(GOROOT)/src/Make.cmd\n")
		_, err = fmt.Fprintf(file, "\n")
		relCmd := path.Join("$(GBROOT)", GetBuildDirCmd())
		_, err = fmt.Fprintf(file, "# gb: copy to local install\n")
		_, err = fmt.Fprintf(file, "%s/$(TARG): $(TARG)\n", relCmd)
		_, err = fmt.Fprintf(file, "\tmkdir -p $(dir $@); cp -f $< $@\n")
		_, err = fmt.Fprintf(file, "command: %s/$(TARG)\n\n", relCmd)
		_, err = fmt.Fprintf(file, "\n")
		if len(this.DepPkgs) != 0 {
			_, err = fmt.Fprintf(file, "# gb: local dependencies\n")
		}
		for _, pkg := range this.DepPkgs {
			_, err = fmt.Fprintf(file, "$(TARG): %s/%s.a\n", relObj, pkg.Target)
		}
	} else {
		_, err = fmt.Fprintf(file, "# gb: copy to local install\n")
		_, err = fmt.Fprintf(file, "%s/$(TARG).a: _obj/$(TARG).a\n", relObj)
		_, err = fmt.Fprintf(file, "\tmkdir -p $(dir $@); cp -f $< $@\n")
		_, err = fmt.Fprintf(file, "package: %s/$(TARG).a\n\n", relObj)
		_, err = fmt.Fprintf(file, "include $(GOROOT)/src/Make.pkg\n")
		_, err = fmt.Fprintf(file, "\n")
		if len(this.DepPkgs) != 0 {
			_, err = fmt.Fprintf(file, "# gb: local dependencies\n")
		}
		for _, pkg := range this.DepPkgs {
			_, err = fmt.Fprintf(file, "%s/$(TARG).a: %s/%s.a\n", GetBuildDirPkg(), relObj, pkg.Target)
		}
	}

	err = file.Close()

	return
}

func (this *Package) AddToBuild(bfile *os.File) {
	if this.addedToBuild {
		return
	}
	this.addedToBuild = true

	if Exclusive && !ListedDirs[this.Dir] {
		return
	}

	if !this.Active {
		return
	}

	for _, pkg := range this.DepPkgs {
		pkg.AddToBuild(bfile)
	}
	fmt.Fprintf(bfile, "&& echo \"(in %s)\" && cd %s && make $1 && cd - > /dev/null \\\n", this.Dir, this.Dir)
}

func (this *Package) GoFMT() (err os.Error) {
	if this.gofmted || (Exclusive && !ListedDirs[this.Dir]) {
		return
	}

	if !this.Active {
		return
	}

	this.gofmted = true

	for _, pkg := range this.DepPkgs {
		if Concurrent {
			go pkg.GoFMT()
		} else {
			err = pkg.GoFMT()
			if err != nil {
				return
			}
		}
	}

	fmt.Printf("(in %s) running gofmt\n", this.Dir)
	for _, src := range this.Sources {
		err = RunGoFMT(this.Dir, src)
		if err != nil {
			return
		}
	}

	this.Stat()

	return
}
