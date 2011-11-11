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
	"testing"
)

type GATest struct {
	path, cwd string
	truth     string
}

func TestGetAbs(t *testing.T) {
	gaTests := []GATest{
		{`package2`, `/home/user/workspace`, `/home/user/workspace/package2`},
		{`/home/go`, `/home/user/workspace`, `/home/go`},
	}
	gaTestsWindows := []GATest{
		{`C:\something`, `/home/user/workspace`, `C:/something`},
		{`C:/something`, `/home/user/workspace`, `C:/something`},
		{`A:\something/else`, `/home/user/workspace`, `A:/something/else`},
	}

	for _, gat := range gaTests {
		result := GetAbs(gat.path, gat.cwd)
		if result != gat.truth {
			t.Error(fmt.Sprintf("GetAbs(\"%s\", \"%s\") -> %s, was expecting %s", gat.path, gat.cwd, result, gat.truth))
		}
	}

	TestWindows = true
	for _, gat := range gaTestsWindows {
		result := GetAbs(gat.path, gat.cwd)
		if result != gat.truth {
			t.Error(fmt.Sprintf("GetAbs(\"%s\", \"%s\") -> %s, was expecting %s", gat.path, gat.cwd, result, gat.truth))
		}
	}
	TestWindows = false
}

type GRTest struct {
	start, finish, wd string
	truth             string
}

func TestGetRelative(t *testing.T) {
	grTests := []GRTest{
		{`/home/projects/goproject`, `/home/go`, `these_are_absolute`, `../../go`},
		{`/home/go`, `/home/go/src/pkg/project`, `these_are_absolute`, `src/pkg/project`},
		{`/home/go`, `package2`, `/home/user/workspace`, `../user/workspace/package2`},
		{`../dir1`, `../dir2`, `/home/user/workspace`, `../dir2`},
		{`a/b/c`, `a/b/cde`, `/home`, `../cde`},
	}
	grTestsWindows := []GRTest{
		{`C:/c/go/gc`, `C:\c\go\go-gb\example`, `wd_does_not_matter_here`, `../go-gb/example`},
		{`C:/a/b/c`, `D:/e/f/g`, `E:/1/2/3`, `D:/e/f/g`},
		{`C:\a\b/c`, `D:/e\f/g`, `E:/1/2/3`, `D:/e/f/g`},
		{`C:/a/b/c`, `D:/e/f/g`, `no_wd`, `D:/e/f/g`},
		{`D:\e\f\g`, `a/b/c`, `D:/home`, `../../../home/a/b/c`},
		{`D:\e\f\g`, `a/b/c`, `e:\\home`, `E:/home/a/b/c`},
		{`D:\e\f\g`, `a/b/c`, `e:/home`, `E:/home/a/b/c`},
		{`e:\dnload\go-lang\go07\go`, `.`, `E:\prog\splitsound\repo01`, `../../../../prog/splitsound/repo01`},
		{`A:\go\walk\src\pkg\walk`, `C:\go\src\pkg`, `A:\go\walk`, `C:/go/src/pkg`},
	}

	for _, grt := range grTests {
		result := GetRelative(grt.start, grt.finish, grt.wd)
		if result != grt.truth {
			t.Error(fmt.Sprintf("GetRelative(\"%s\", \"%s\", \"%s\") -> \"%s\", was expecting \"%s\"", grt.start, grt.finish, grt.wd, result, grt.truth))
		}
	}

	TestWindows = true
	for _, grt := range grTestsWindows {
		result := GetRelative(grt.start, grt.finish, grt.wd)
		if result != grt.truth {
			t.Error(fmt.Sprintf("GetRelative(\"%s\", \"%s\", \"%s\") -> \"%s\", was expecting \"%s\"", grt.start, grt.finish, grt.wd, result, grt.truth))
		}
	}
	TestWindows = false
}

func BenchmarkX(b *testing.B) {
	//do nothing
}
