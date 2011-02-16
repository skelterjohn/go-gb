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

func Usage() {
	fmt.Printf("Usage: gb [-options] [directory list]\n")
	fmt.Printf("Options:\n")
	fmt.Printf(" ? print this usage text\n")
	fmt.Printf(" i install\n")
	fmt.Printf(" c clean\n")
	fmt.Printf(" N nuke\n")
	fmt.Printf(" b build after cleaning\n")
	fmt.Printf(" g use goinstall when appropriate\n")
	fmt.Printf(" G use goinstall -u when possible\n")
	fmt.Printf(" p build packages in parallel, when possible\n")
	fmt.Printf(" s scan and list targets without building\n")
	fmt.Printf(" S scan and list targets and their dependencies without building\n")
	fmt.Printf(" t run tests\n")
	fmt.Printf(" e exclusive target list (do not build/clean/test/install a target unless it\n")
	fmt.Printf("   resides in a listed directory)\n")
	fmt.Printf(" v verbose\n")
	fmt.Printf(" m use makefiles, when possible\n")
	fmt.Printf(" M generate standard makefiles without building\n")
	fmt.Printf(" f force overwrite of existing makefiles\n")
	fmt.Printf(" F run gofmt on source files in targeted directories\n")
	fmt.Printf(" P build/clean/install only packages\n")
	fmt.Printf(" C build/clean/install only cmds\n")
	fmt.Printf(" D create distribution\n")
	fmt.Printf(" W create workspace.gb files in all directories\n")
	fmt.Printf(" R update dependencies in $GOROOT/src\n")
}
