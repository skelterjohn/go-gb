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
	"exec"
	"os"
	"fmt"
)

func MakeBuild(pkg *Package) (err os.Error) {
	buildBlock <- true
	defer func() { <-buildBlock }()
	
	if pkg.NeedsBuild {
		MakeClean(pkg)
	}
	
	margs := []string{"make"}
	if Install || pkg.IsInGOROOT {
		margs = append(margs, "install")
	}
	//fmt.Printf("(in %v)\n", pkg.Dir)
	fmt.Printf("%v\n", margs)
	p, err := exec.Run(MakeCMD, margs, os.Envs, pkg.Dir, exec.PassThrough, exec.PassThrough, exec.PassThrough)
	if err != nil {
		return
	}
	if p != nil {
		p.Wait(0)
	}
	return
}

func MakeClean(pkg *Package) (err os.Error) {
	margs := []string{"make", "clean"}
	fmt.Printf("(in %v)\n", pkg.Dir)
	fmt.Printf("%v\n", margs)
	p, err := exec.Run(MakeCMD, margs, os.Envs, pkg.Dir, exec.PassThrough, exec.PassThrough, exec.PassThrough)
	if err != nil {
		return
	}
	if p != nil {
		p.Wait(0)
	}
	return
}

func MakeTest(pkg *Package) (err os.Error) {
	margs := []string{"make", "test"}
	fmt.Printf("(in %v)\n", pkg.Dir)
	fmt.Printf("%v\n", margs)
	p, err := exec.Run(MakeCMD, margs, os.Envs, pkg.Dir, exec.PassThrough, exec.PassThrough, exec.PassThrough)
	if err != nil {
		return
	}
	if p != nil {
		p.Wait(0)
	}
	return
}
