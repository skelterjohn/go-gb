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
	"path"
	"regexp"
	"strings"
)

var disabledGCRE = regexp.MustCompile(`^([a-z0-9\-]+\.googlecode\.com/(svn|hg))(/[a-z0-9A-Z_.\-/]*)?$`)

//taken from goinstall source
var goinstallables = []*regexp.Regexp{
	regexp.MustCompile(`^([a-z0-9\-]+\.googlecode\.com/(svn|hg))(/[a-z0-9A-Z_.\-/]*)?$`),
	regexp.MustCompile(`^code\.google\.com/p/([a-z0-9\-]+(\.[a-z0-9\-]+)?)(/[a-z0-9A-Z_.\-/]+)?$`),
	regexp.MustCompile(`^(github\.com/[a-z0-9A-Z_.\-]+/[a-z0-9A-Z_.\-]+)(/[a-z0-9A-Z_.\-/]*)?$`),
	regexp.MustCompile(`^(bitbucket\.org/[a-z0-9A-Z_.\-]+/[a-z0-9A-Z_.\-]+)(/[a-z0-9A-Z_.\-/]*)?$`),
	regexp.MustCompile(`^(launchpad\.net/([a-z0-9A-Z_.\-]+(/[a-z0-9A-Z_.\-]+)?|~[a-z0-9A-Z_.\-]+/(\+junk|[a-z0-9A-Z_.\-]+)/[a-z0-9A-Z_.\-]+))(/[a-z0-9A-Z_.\-/]+)?$`),
	regexp.MustCompile(`.+{\.hg|\.git|\.bzr|\.svn}[/.*]`),
}

var goinstalledAlready = make(map[string]bool)

func IsGoInstallable(target string) (matches bool) {
	target = strings.Trim(target, "\"")

	for _, re := range goinstallables {
		if m := re.FindStringSubmatch(target); m != nil {
			matches = true
			break
		}
	}

	return
}

func GoInstallPkg(target string) (touched int64) {
	if goinstalledAlready[target] {
		return
	}
	goinstalledAlready[target] = true

	if disabledGCRE.FindStringSubmatch(target) != nil {
		WarnLog.Printf("Googlecode format \"%s\" is no longer accepted - use gofix")
		return
	}

	if !IsGoInstallable(target) {
		return
	}

	target = strings.Trim(target, "\"")

	argv := []string{"goinstall", target}
	if GoInstallUpdate {
		argv = []string{"goinstall", "-u", "-clean", target}
	}

	fmt.Printf("%v\n", argv)

	err := RunExternal(GoInstallCMD, ".", argv)
	if err != nil {
		return
	}

	goinstalledFile := path.Join(GetInstallDirPkg(), target) + ".a"

	touched, _ = StatTime(goinstalledFile)
	return
}
