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
	"strings"
	"path"
)

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
	p = path.Clean(p)
	// Work around IsAbs() not working on windows
	if GOOS == "windows" {
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
	abspath = path.Clean(path.Join(wd, p))
	return
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
	for !(strings.HasPrefix(finish, start+path.DirSeps) || start == "/") {
		backtracking = path.Join(backtracking, "..")
		start, _ = path.Split(start)
		start = path.Clean(start)
	}
	if start == finish {
		return path.Clean(path.Join(backtracking, ".")), nil
	}
	if start == "/" {
		relative = path.Clean(path.Join(backtracking, finish))
	} else {
		relative = path.Clean(path.Join(backtracking, finish[len(start)+1:len(finish)]))
	}
	return
}
