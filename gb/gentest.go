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
	"template"
)

type TestPkg struct {
	PkgAlias, PkgName, PkgTarget string
	TestFuncs, TestBenchmarks    []string
}

type TestSuite struct {
	TestPkgs []*TestPkg
}

var TestmainTemplate = template.Must(template.New("TestSource").Parse(
	`package main

{{range .TestPkgs}}import {{.PkgAlias}} "{{.PkgTarget}}"
{{end}}
import "testing"
import "os"
import __regexp__ "regexp"

var tests = []testing.InternalTest{
{{range .TestPkgs}}{{if $PkgName=.PkgName}}{{if $PkgAlias=.PkgAlias}}{{range .TestFuncs}}	{"{{$PkgName}}.{{.}}", {{$PkgAlias}}.{{.}}},{{end}}{{end}}{{end}}{{end}}
}

var benchmarks = []testing.InternalBenchmark{
{{range .TestPkgs}}{{if $PkgName=.PkgName}}{{if $PkgAlias=.PkgAlias}}{{range .TestBenchmarks}}	{"{{$PkgName}}.{{.}}", {{$PkgAlias}}.{{.}}},{{end}}{{end}}{{end}}{{end}}
}

var matchPat string
var matchRe *__regexp__.Regexp

func matchString(pat, str string) (result bool, err os.Error) {
	if matchRe == nil || matchPat != pat {
		matchPat = pat
		matchRe, err = __regexp__.Compile(matchPat)
		if err != nil {
			return
		}
	}
	return matchRe.MatchString(str), nil
}

func main() {
	testing.Main(matchString, tests, benchmarks)
}
`))
