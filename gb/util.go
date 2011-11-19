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
	"path/filepath"
	"runtime"
	"strings"
)

var TestWindows = false

// GetAbs returns the absolute version of the path supplied.
func GetAbs(p, cwd string) (abspath string) {
	p = pathClean(p)
	cwd = pathClean(cwd)
	// Work around IsAbs() not working on windows
	if TestWindows || runtime.GOOS == "windows" {
		if len(p) > 2 && p[1:3] == ":/" {
			abspath = p
			return
		}
	} else {
		if filepath.IsAbs(p) {
			abspath = p
			return
		}
	}
	abspath = path.Join(cwd, p)
	return
}

func GetRoot(p string) (r string) {
	if (TestWindows || runtime.GOOS == "windows") && len(p) > 1 && p[1] == ':' {
		return p[0:2] + "/"
	}
	return "/"
}

func pathClean(p string) (r string) {
	if TestWindows || runtime.GOOS == "windows" {
		p = strings.Replace(p, "\\", "/", -1)
		if len(p) >= 2 && p[1] == ':' {
			p = strings.ToUpper(p[0:1]) + p[1:len(p)]
		}
	}
	r = path.Clean(p)
	return
}

func HasPathPrefix(p, pr string) bool {
	p = pathClean(p)
	pr = pathClean(pr)

	if pr == GetRoot(p) {
		return true
	}

	if len(pr) == 0 {
		return false
	}
	if len(pr) > len(p) {
		return false
	}
	if len(p) == len(pr) {
		return p == pr
	}
	if pr[len(pr)-1] == '/' {
		return p[0:len(pr)] == pr
	}
	return p[0:len(pr)+1] == pr+"/"
}

// GetRelative(start, finish) returns the path to finish, relative to start.
func GetRelative(start, finish, cwd string) (relative string) {
	start = GetAbs(pathClean(start), cwd)
	finish = GetAbs(pathClean(finish), cwd)
	cwd = pathClean(cwd)

	if TestWindows || runtime.GOOS == "windows" {
		if len(start) >= 2 && len(finish) >= 2 {
			if start[0] != finish[0] && start[1] == finish[1] {
				relative = finish //absolute path is the only way
				return
			}
		}
	}

	backtracking := "."
	for !HasPathPrefix(finish, start) {
		backtracking = path.Join(backtracking, "..")
		start, _ = path.Split(start)
		start = pathClean(start)
	}
	if start == finish {
		return pathClean(path.Join(backtracking, "."))
	}
	if start == "/" {
		relative = path.Join(backtracking, finish)
	} else {
		relative = path.Join(backtracking, finish[len(start)+1:len(finish)])
	}
	return
}
