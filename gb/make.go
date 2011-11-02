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
)

func MakeBuild(pkg *Package) (err error) {
	if MakeCMD == "" {
		fmt.Printf("(in %s) Can't use make to build %s\n", pkg.Dir, pkg.Target)
		return
	}

	margs := []string{"gomake", "clean"}
	if Install || pkg.IsInGOROOT {
		margs = append(margs, "install")
	} else {
		if pkg.IsCmd {
			margs = append(margs, "command")
		} else {
			margs = append(margs, "package")
		}
	}
	//fmt.Printf("(in %v)\n", pkg.Dir)
	fmt.Printf("%v\n", margs)
	err = RunExternal(MakeCMD, pkg.Dir, margs)
	return
}

func MakeClean(pkg *Package) (err error) {
	if MakeCMD == "" {
		fmt.Printf("(in %s) Can't use make to clean %s\n", pkg.Dir, pkg.Target)
		return
	}
	margs := []string{"gomake", "clean"}
	if Nuke {
		margs = append(margs, "nuke")
	}
	fmt.Printf("(in %v)\n", pkg.Dir)
	fmt.Printf("%v\n", margs)
	err = RunExternal(MakeCMD, pkg.Dir, margs)
	return
}

func MakeTest(pkg *Package) (err error) {
	margs := []string{"gomake", "test"}
	fmt.Printf("(in %v)\n", pkg.Dir)
	fmt.Printf("%v\n", margs)
	err = RunExternal(MakeCMD, pkg.Dir, margs)
	return
}
