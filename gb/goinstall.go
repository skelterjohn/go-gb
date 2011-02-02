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
	"regexp"
	"fmt"
	"exec"
	"os"
	"path"
)

//taken from goinstall source
var googlecode = regexp.MustCompile(`^([a-z0-9\-]+\.googlecode\.com/(svn|hg))(/[a-z0-9A-Z_.\-/]*)?$`)
var github = regexp.MustCompile(`^(github\.com/[a-z0-9A-Z_.\-]+/[a-z0-9A-Z_.\-]+)(/[a-z0-9A-Z_.\-/]*)?$`)
var bitbucket = regexp.MustCompile(`^(bitbucket\.org/[a-z0-9A-Z_.\-]+/[a-z0-9A-Z_.\-]+)(/[a-z0-9A-Z_.\-/]*)?$`)
var launchpad = regexp.MustCompile(`^(launchpad\.net/([a-z0-9A-Z_.\-]+(/[a-z0-9A-Z_.\-]+)?|~[a-z0-9A-Z_.\-]+/(\+junk|[a-z0-9A-Z_.\-]+)/[a-z0-9A-Z_.\-]+))(/[a-z0-9A-Z_.\-/]+)?$`)

var goinstallables = []*regexp.Regexp{googlecode, github, bitbucket, launchpad}

var goinstallBlock = make(chan bool, 1)

var goinstalledAlready = make(map[string]bool)

func IsGoInstallable(target string) (matches bool) {
	//trim quote marks
	if target[0] == '"' {
		target = target[1:len(target)]
	}
	if target[len(target)-1] == '"' {
		target = target[0 : len(target)-1]
	}

	for _, re := range goinstallables {
		if m := re.FindStringSubmatch(target); m != nil {
			matches = true
			break
		}
	}

	return
}

func GoInstallPkg(target string) (touched int64) {
	goinstallBlock <- true
	defer func() { <-goinstallBlock }()

	if goinstalledAlready[target] {
		return
	}
	goinstalledAlready[target] = true

	if !IsGoInstallable(target) {
		return
	}

	//trim quote marks
	if target[0] == '"' {
		target = target[1:len(target)]
	}
	if target[len(target)-1] == '"' {
		target = target[0 : len(target)-1]
	}

	argv := []string{"goinstall", target}
	if GoInstallUpdate {
		argv = []string{"goinstall", "-u", target}
	}
	//if Verbose {
	fmt.Printf("%v\n", argv)
	//}
	p, err := exec.Run(GoInstallCMD, argv, os.Envs, ".", exec.PassThrough, exec.PassThrough, exec.PassThrough)
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

	goinstalledFile := path.Join(GetInstallDirPkg(), target) + ".a"

	var info *os.FileInfo
	info, err = os.Stat(goinstalledFile)
	if err != nil {
		return
	}
	return info.Mtime_ns

}
