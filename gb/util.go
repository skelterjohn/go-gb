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
	"fmt"
	"path"
)

func GetAbsolutePath(p string) (absp string, err os.Error) {
	if path.IsAbs(p) {
		absp = p
		return
	}
	wd, err := os.Getwd()
	if p == "." {
		absp = path.Clean(wd)
		return
	}
	absp = path.Join(wd, p)
	return
}

func GetRelativePath(parent, child string) (rel string, err os.Error) {
	parent, err = GetAbsolutePath(parent)
	child, err = GetAbsolutePath(child)

	if !strings.HasPrefix(child, parent) {
		err = os.NewError(fmt.Sprintf("'%s' is not in '%s'", child, parent))
		return
	}

	rel = path.Clean(child[len(parent)+1 : len(child)])

	return
}
