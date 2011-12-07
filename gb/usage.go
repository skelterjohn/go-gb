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

var UsageText = `Usage: gb [options] [directory list]
Options:
 -b build after cleaning
 -c clean
 -C build/clean/install only cmds
 -e exclusive target list (do not build/clean/test/install a target unless it
    resides in a listed directory)
 -f force overwrite of existing makefiles
 -g use goinstall when appropriate
 -G use "goinstall -clean -u" when possible
 -h print this usage text
 -i install
 -L scan and list targets and their source files
 -m use makefiles, when possible
 -N nuke
 -p build packages in parallel, when possible
 -P build/clean/install only packages
 -R update dependencies in $GOROOT/src
 -s scan and list targets without building
 -S scan and list targets and their dependencies without building
 -t run tests
 -v verbose
 --gofmt
     run gofmt on source files in targeted directories
 --makefiles
     generate standard makefiles without building
 --workspace
     create workspace.gb files in all directories
 --testargs
     all arguments following --testargs are passed to the test binary
`

func Usage() {
	fmt.Print(UsageText)
}
