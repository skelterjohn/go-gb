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
	"exec"
	"fmt"
)

var MakeCMD, CompileCMD, LinkCMD, PackCMD, CopyCMD, GoInstallCMD, GoFMTCMD string

func FindExternals() (err os.Error) {
	var err2 os.Error
	MakeCMD, err2 = exec.LookPath("make")
	if err2 != nil {
		fmt.Printf("Could not find 'make' in path\n")
	}

	CompileCMD, err = exec.LookPath(GetCompilerName())
	if err != nil {
		fmt.Printf("Could not find '%s' in path\n", GetCompilerName())
		return
	}

	LinkCMD, err = exec.LookPath(GetLinkerName())
	if err != nil {
		fmt.Printf("Could not find '%s' in path\n", GetLinkerName())
		return
	}
	PackCMD, err = exec.LookPath("gopack")
	if err != nil {
		fmt.Printf("Could not find 'gopack' in path\n")
		return
	}
	CopyCMD, _ = exec.LookPath("cp")

	GoInstallCMD, err2 = exec.LookPath("goinstall")
	if err != nil {
		fmt.Printf("Could not find 'goinstall' in path\n")
	}
	GoFMTCMD, err2 = exec.LookPath("gofmt")
	if err != nil {
		fmt.Printf("Could not find 'gofmt' in path\n")
	}

	return
}

func RunExternal(cmd, wd string, argv []string) (err os.Error) {
	var p *exec.Cmd
	p, err = exec.Run(cmd, argv, os.Envs, wd, exec.PassThrough, exec.PassThrough, exec.PassThrough)
	if err != nil {
		return
	}
	if p != nil {
		var wmsg *os.Waitmsg
		wmsg, err = p.Wait(0)
		if wmsg.ExitStatus() != 0 {
			err = os.NewError(fmt.Sprintf("%v: %s\n", argv, wmsg.String()))
			return
		}
		if err != nil {
			return
		}
	}
	return
}
