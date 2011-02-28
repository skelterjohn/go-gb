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

var UsageText = `Usage: gb [-options] [directory list]
Options:
 ? print this usage text
 i install
 c clean
 N nuke
 b build after cleaning
 g use goinstall when appropriate
 G use goinstall -u when possible
 p build packages in parallel, when possible
 s scan and list targets without building
 S scan and list targets and their dependencies without building
 L scan and list targets and their source files
 t run tests
 e exclusive target list (do not build/clean/test/install a target unless it
   resides in a listed directory)
 v verbose
 m use makefiles, when possible
 M generate standard makefiles without building
 f force overwrite of existing makefiles
 F run gofmt on source files in targeted directories
 P build/clean/install only packages
 C build/clean/install only cmds
 D create distribution
 W create workspace.gb files in all directories
 R update dependencies in $GOROOT/src
`

func Usage() {
	fmt.Printf(UsageText)
}
