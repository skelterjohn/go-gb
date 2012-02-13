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
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var MakeCMD,
	CompileCMD,
	CCMD,
	AsmCMD,
	LinkCMD,
	PackCMD,
	CopyCMD,
	GoInstallCMD,
	GoFMTCMD,
	GoFixCMD,
	GoCMD,
	CGoCMD,
	GCCCMD,
	ProtocCMD,
	GoYaccCMD string

func FindGobinExternal(name string) (path string, err error) {
	path, err = exec.LookPath(name)
	if err != nil {
		path = filepath.Join(GOBIN, name)
		_, err = os.Stat(path)
	}
	if err != nil {
		path = filepath.Join(GOBIN, "tool", name)
		_, err = os.Stat(path)
	}
	return
}

func FindExternals() (err error) {
	GoCMD, err = FindGobinExternal("go")
	if err != nil {
		fmt.Printf("Could not find 'go' in path\n")
		return
	}

	CompileCMD = "go tool " + GetCompilerName()
	AsmCMD = "go tool " + GetAssemblerName()
	LinkCMD = "go tool " + GetLinkerName()
	PackCMD = "go tool pack"
	CGoCMD = "go tool cgo"
	GoFMTCMD = "go fmt"
	GoFixCMD = "go fix"
	GCCCMD, _ = exec.LookPath("gcc")
	CCMD = "go tool " + GetCCompilerName()
	GoYaccCMD = "go tool yacc"

	ProtocCMD, _ = exec.LookPath("protoc")

	CopyCMD, _ = exec.LookPath("cp")

	return
}

func SplitArgs(args []string) (sargs []string) {
	for _, arg := range args {
		sarg := strings.Fields(arg)
		sargs = append(sargs, sarg...)
	}
	return
}

func RunExternalDump(cmd, wd string, argv []string, dump *os.File) (err error) {
	argv = SplitArgs(argv)

	if strings.Index(cmd, " ") != -1 {
		cmds := strings.Fields(cmd)
		argv = append(cmds, argv[1:]...)

	}

	if argv[0] == "go" {
		cmd = GoCMD
	}
	if Verbose {
		fmt.Printf("%s\n", argv)
	}

	c := exec.Command(cmd, argv[1:]...)
	c.Dir = wd
	c.Env = os.Environ()

	c.Stdout = dump
	c.Stderr = os.Stderr

	err = c.Run()

	if wmsg, ok := err.(*exec.ExitError); ok {
		if wmsg.ExitStatus() != 0 {
			err = errors.New(fmt.Sprintf("%v: %s\n", argv, wmsg.String()))
		} else {
			err = nil
		}
	}
	return
}
func RunExternal(cmd, wd string, argv []string) (err error) {
	return RunExternalDump(cmd, wd, argv, os.Stdout)
}
