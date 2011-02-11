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
	"os"
	"runtime"
	"path"
	"strings"
)

var TestWindows = false

func StatTime(p string) (time int64, err os.Error) {
	var info *os.FileInfo
	info, err = os.Stat(p)
	if err != nil {
		return
	}
	time = info.Mtime_ns
	return
}

// GetAbs returns the absolute version of the path supplied.
func GetAbs(p string) (abspath string, err os.Error) {
	p = pathClean(p)
	// Work around IsAbs() not working on windows
	if (TestWindows || runtime.GOOS == "windows") {
		if len(p) > 1 && p[1] == ':' {
			abspath = p
			return
		}
	} else {
		if path.IsAbs(p) {
			abspath = p
			return
		}
	}
	var wd string
	wd, err = os.Getwd()
	abspath = path.Join(wd, p)
	return
}

func GetRoot(p string) (r string) {
	if (TestWindows || runtime.GOOS == "windows") && len(p) > 1 && p[1] == ':' {
		return p[0:2]+path.DirSeps
	} 
	return "/"
}

func pathClean(p string) (r string) {
	if (TestWindows || runtime.GOOS == "windows") {
		p = strings.Replace(p, "\\", "/", -1)
		if len(p)>=2 && p[1] == ':' {
			p = strings.ToUpper(p[0:1])+p[1:len(p)]
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
	if p == pr {
		return true
	}
	if pr[len(pr)-1] == path.DirSeps[0] {
		return p[0:len(pr)] == pr
	}
	return p[0:len(pr)+1] == pr+path.DirSeps
}

// GetRelative(start, finish) returns the path to finish, relative to start.
func GetRelative(start, finish string) (relative string, err os.Error) {
	if start, err = GetAbs(start); err != nil {
		return
	}
	if finish, err = GetAbs(finish); err != nil {
		return
	}
	
	backtracking := "."
	for !HasPathPrefix(finish, start) {
		backtracking = path.Join(backtracking, "..")
		start, _ = path.Split(start)
		start = path.Clean(start)
	}
	if start == finish {
		return pathClean(path.Join(backtracking, ".")), nil
	}
	if start == "/" {
		relative = path.Join(backtracking, finish)
	} else {
		relative = path.Join(backtracking, finish[len(start)+1:len(finish)])
	}
	return
}
