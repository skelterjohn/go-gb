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
	Dir         string
	Name        string
	Target      string
	Base        string
	IsCmd       bool
	Sources     []string
	Deps        []string
	ib          string
	result      string
	installPath string
	MyErr       os.Error
	Active      bool

	built, cleaned, addedToBuild, gofmted bool

	CGoSources []string
	CSrcs      []string

	TestSources []string
	TestDeps    []string

	Funcs []string

	HasMakefile bool

	DepPkgs     []*Package
	TestDepPkgs []*Package

	SourceTime, BinTime, InstTime int64

	block chan bool
}

func ReadPackage(base, dir string) (this *Package, err os.Error) {
	finfo, err := os.Stat(dir)
	if err != nil || !finfo.IsDirectory() {
		err = os.NewError("not a directory")
		return
	}
	this = new(Package)
	this.block = make(chan bool, 1)
	this.Dir = path.Clean(dir)
	
	
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
		var srcInfo *os.FileInfo
		srcInfo, err = os.Stat(path.Join(this.Dir, src))
		if err != nil {
			return
		}
		t := srcInfo.Mtime_ns
		if t > this.SourceTime {
			this.SourceTime = t
		}
	}

	if err != nil {
		return
	}
	this.IsCmd = this.Name == "main"
	this.ib = path.Join(this.Dir, GetIBName())
	err = this.GetTarget()

	this.Active = (DoCmds && this.IsCmd) || (DoPkgs && !this.IsCmd)


	return
}

func (this *Package) ScanForSource() (err os.Error) {
	errch := make(chan os.Error)
	path.Walk(this.Dir, this, errch)
	
	if len(this.Sources) + len(this.TestSources) == 0 {
		err = os.NewError("No source files in " + this.Dir)
	}
	
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
	rootl := len(this.Dir)+1
	if this.Dir != "." {
		fpath = fpath[rootl:len(fpath)]
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
	for _, src := range this.Sources {
		var fpkg, ftarget string
		var fdeps []string
		fpkg, ftarget, fdeps, _, err = GetDeps(path.Join(this.Dir, src))
		if this.Name != "" && fpkg != this.Name {
			err = os.NewError(fmt.Sprintf("in %s: Source for more than one target", this.Dir))
			fmt.Printf("%v\n", err)
			return
		}
		if err != nil {
			return
		}
		if ftarget != "" {
			this.Target = ftarget
		}
		this.Name = fpkg
		this.Deps = append(this.Deps, fdeps...)
	}
	this.Deps = RemoveDups(this.Deps)
	if Test {
		for _, src := range this.TestSources {
			var fpkg, ftarget string
			var fdeps, ffuncs []string
			fpkg, ftarget, fdeps, ffuncs, err = GetDeps(path.Join(this.Dir, src))
			if this.Name != "" && fpkg != this.Name {
				err = os.NewError(fmt.Sprintf("in %s: more than one test package (ignoring test source)", this.Dir))
				fmt.Printf("%v\n", err)
				this.TestDeps = nil
				break
			}
			if err != nil {
				break
			}
			if ftarget != "" {
				this.Target = ftarget
			}
			this.Name = fpkg
			this.TestDeps = append(this.TestDeps, fdeps...)
			this.Funcs = append(this.Funcs, ffuncs...)
		}
		this.TestDeps = RemoveDups(this.TestDeps)
	}
	return
	//fpkg, ftarget, fdeps, ffuncs, err = GetDeps(srcloc)
}	

func (this *Package) GetTarget() (err os.Error) {
	if this.Target == "" {
		this.Target = this.Base

		if this.IsCmd {
			this.Target = path.Base(this.Dir)
			if this.Target == "." {
				this.Target = "main"
			}
		}
	} else {
		this.Base = this.Target
	}

	tpath := path.Join(this.Dir, "/target.gb")
	fin, err := os.Open(tpath, os.O_RDONLY, 0)
	if err == nil {
		bfrd := bufio.NewReader(fin)
		this.Target, err = bfrd.ReadString('\n')
		this.Target = strings.TrimSpace(this.Target)
		this.Base = this.Target
		if this.Target == "-" {
			err = os.NewError("directory opts-out")
			return
		}
	}

	this.Base = path.Clean(this.Base)
	this.Target = path.Clean(this.Target)

	err = nil
	if this.IsCmd {
		if GOOS == "windows" {
			this.Target += ".exe"
		}
		this.installPath = path.Join(GetInstallDirCmd(), this.Target)
		this.result = path.Join(GetBuildDirCmd(), this.Target)
	} else {

		this.installPath = path.Join(GetInstallDirPkg(), this.Target+".a")
		this.result = path.Join(GetBuildDirPkg(), this.Target+".a")
	}

	this.Stat()

	return
}

func (this *Package) Stat() {
	resInfo, err := os.Stat(this.result)
	if resInfo != nil && err == nil {
		this.BinTime = resInfo.Mtime_ns
	} else {
		this.BinTime = 0
	}
	resInfo, err = os.Stat(this.installPath)
	if resInfo != nil && err == nil {
		this.InstTime = resInfo.Mtime_ns
	} else {
		this.InstTime = 0
	}
}

func (this *Package) ResolveDeps() (err os.Error) {
	CheckDeps := func(deps []string) (err os.Error) {
		for _, dep := range deps {
			if pkg, ok := Packages[dep]; ok {
				this.DepPkgs = append(this.DepPkgs, pkg)
			} else if !IsGoInstallable(dep) {
				if !PkgExistsInGOROOT(dep) {
					fmt.Printf("in %s: can't resolve pkg %s (maybe you aren't in the root?)\n", this.Dir, dep)
					err = os.NewError("unresolved packages")
				}
			} else {
				if !PkgExistsInGOROOT(dep) && !GoInstall {
					fmt.Pprintf("in %s: can't resolve pkg %s (try using -g)\n", this.Dir, dep)
					err = os.NewError("unresolved packages")
				}
			}
		}
		return
	}
	err = CheckDeps(this.Deps)
	if err != nil {
		return
	}
	err = CheckDeps(this.TestDeps)
	return
}

func (this *Package) Touched() (build, install bool) {

	var inTime int64

	for _, pkg := range this.DepPkgs {
		if pkg.BinTime > inTime {
			inTime = pkg.BinTime
		}
		db, di := pkg.Touched()
		if db {
			build = true
		}
		if di {
			install = true
		}
	}
	if this.SourceTime > inTime {
		inTime = this.SourceTime
	}
	if inTime > this.BinTime {
		build = true
	}
	if this.InstTime < this.BinTime {
		install = true
	}
	
	if build == true {
		install = true
	}
	
	return
}

func (this *Package) Build() (err os.Error) {
	if this.built {
		return
	}
	this.built = true
	if this.MyErr != nil {
		return
	}
	if !this.HasMakefile && len(this.CGoSources) + len(this.CSrcs) != 0 {
		fmt.Printf("(in %s) this is a cgo project; please create a makefile", this.Dir)
		return
	}
	this.block <- true
	defer func() {
		this.MyErr = err
		if this.MyErr != nil {
			BrokenPackages++
		}
		<-this.block
	}()

	if Exclusive && !ListedDirs[this.Dir] {
		return
	}

	if Concurrent {
		for _, pkg := range this.DepPkgs {
			go pkg.Build()
		}
	}

	var inTime int64

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
		if Makefiles && this.HasMakefile || (len(this.CGoSources) + len(this.CSrcs) != 0) { 
			err = MakeBuild(this)
		} else {
			err = BuildPackage(this)
		}
		if err == nil {
			PackagesBuilt++
		}
	}

	if err != nil {
		this.CleanFiles()
	}

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

	if Makefiles && this.HasMakefile || (len(this.CGoSources) + len(this.CSrcs) != 0)  {
		err = MakeTest(this)
		return
	}

	testdir := path.Join(this.Dir, "_test")
	if Verbose {
		fmt.Printf(" Removing %s\n", testdir)
	}
	err = os.RemoveAll(testdir)

	fmt.Printf("(in %s) testing \"%s\"\n", this.Dir, this.Target)

	var tests, benchmarks []string

	for _, f := range this.Funcs {
		if strings.HasPrefix(f, "Test") {
			tests = append(tests, f)
		}
		if strings.HasPrefix(f, "Benchmark") {
			benchmarks = append(benchmarks, f)
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

	fmt.Fprintf(file, "import \"%s\"\n", this.Target)
	fmt.Fprintf(file, "import \"testing\"\n")
	fmt.Fprintf(file, "import __regexp__ \"regexp\"\n\n")

	fmt.Fprintf(file, "var tests = []testing.InternalTest{\n")
	for _, test := range tests {
		fmt.Fprintf(file, "\t{\"%s.%s\", %s.%s},\n", this.Name, test, this.Name, test)
	}
	fmt.Fprintf(file, "}\n")

	fmt.Fprintf(file, "var benchmarks = []testing.InternalBenchmark{\n")
	for _, benchmark := range benchmarks {
		fmt.Fprintf(file, "\t{\"%s.%s\", %s.%s},\n", this.Name, benchmark, this.Name, benchmark)
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
	if Makefiles && this.HasMakefile || (len(this.CGoSources) + len(this.CSrcs) != 0) {
		MakeClean(this)
		PackagesBuilt++
		return
	}
	ib := false
	res := false
	if _, err2 := os.Stat(this.ib); err2 == nil {
		ib = true
	}
	if _, err2 := os.Stat(this.result); err2 == nil {
		res = true
	}
	if !ib && !res {
		return
	}
	fmt.Printf("Cleaning %s\n", this.Dir)
	if Verbose {
		fmt.Printf(" Removing %s\n", this.ib)
	}
	err = os.Remove(this.ib)
	if Verbose {
		fmt.Printf(" Removing %s\n", this.result)
	}
	err = os.Remove(this.result)
	testdir := path.Join(this.Dir, "_test")
	if Verbose {
		fmt.Printf(" Removing %s\n", testdir)
	}
	err = os.RemoveAll(testdir)

	this.Stat()

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
	if Exclusive && !ListedDirs[this.Dir] {
		return
	}

	for _, pkg := range this.DepPkgs {
		pkg.Install()
	}

	if !this.Active {
		return
	}

	if !(Makefiles && this.HasMakefile) && this.InstTime < this.BinTime {
		err = InstallPackage(this)

		this.Stat()

		PackagesInstalled++
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

	if len(this.CGoSources) != 0 || len(this.CSrcs) != 0 {
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
	_, err = fmt.Fprintf(file, "TARG=%s\n", this.Target)
	_, err = fmt.Fprintf(file, "GOFILES=\\\n")
	for _, src := range this.Sources {
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
		relCmd := path.Join("$(GBROOT)", GetBuildDirCmd())
		_, err = fmt.Fprintf(file, "# gb: copy to local install\n")
		_, err = fmt.Fprintf(file, "%s/$(TARG): $(TARG)\n", relCmd)
		_, err = fmt.Fprintf(file, "\tmkdir -p $(dir $@); cp -f $< $@\n")
		_, err = fmt.Fprintf(file, "command: %s/$(TARG)\n\n", relCmd)
		_, err = fmt.Fprintf(file, "include $(GOROOT)/src/Make.cmd\n")
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
