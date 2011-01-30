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