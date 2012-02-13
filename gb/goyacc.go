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
	"path/filepath"
)

func GoForYacc(yaccsrc string) (gosrc string) {
	gosrc = yaccsrc + ".go"
	return
}

func GenerateGoyaccSource(this *Package) (err error) {
	for _, ys := range this.YaccSrcs {
		base := ys[:len(ys)-len(".y")]
		gosrc := GoForYacc(ys)
		args := []string{"goyacc", "-o", gosrc, "-p", base, ys}

		err = RunExternal(GoYaccCMD, this.Dir, args)
		if err != nil {
			return
		}

		var pkg string
		pkg, _, _, _, _, _, err = GetDeps(filepath.Join(this.Dir, gosrc))
		if err != nil {
			return
		}

		this.PkgSrc[pkg] = append(this.PkgSrc[pkg], gosrc)

		// this probably can't actually happen
		if this.Name == "" {
			this.Name = pkg
		}
	}
	return
}
