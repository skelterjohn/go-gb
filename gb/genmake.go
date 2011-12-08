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
	"template"
)

type MakeData struct {
	Target      string
	GBROOT      string
	GoFiles     []string
	AsmObjs     []string
	CGoFiles    []string
	CObjs       []string
	LocalDeps   []string
	BuildDirPkg string
	BuildDirCmd string
	CopyLocal   bool
}

var MakeCmdTemplate = template.Must(template.New("MakeCmd").Parse(
	`# Makefile generated by gb: http://go-gb.googlecode.com
# gb provides configuration-free building and distributing

include $(GOROOT)/src/Make.inc

TARG={{.Target}}
GOFILES=\
{{range .GoFiles}}	{{.}}\
{{end}}
# gb: this is the local install
GBROOT={{.GBROOT}}

# gb: compile/link against local install
GCIMPORTS+= -I $(GBROOT)/_obj
LDIMPORTS+= -L $(GBROOT)/_obj

# gb: compile/link against GOPATH entries
GOPATHSEP=:
ifeq ($(GOHOSTOS),windows)
GOPATHSEP=;
endif
GCIMPORTS+=-I $(subst $(GOPATHSEP),/pkg/$(GOOS)_$(GOARCH) -I , $(GOPATH))/pkg/$(GOOS)_$(GOARCH)
LDIMPORTS+=-L $(subst $(GOPATHSEP),/pkg/$(GOOS)_$(GOARCH) -L , $(GOPATH))/pkg/$(GOOS)_$(GOARCH)

# gb: default target is in GBROOT this way
command:

include $(GOROOT)/src/Make.cmd

# gb: copy to local install
$(GBROOT)/{{.BuildDirCmd}}/$(TARG): $(TARG)
	mkdir -p $(dir $@); cp -f $< $@
command: $(GBROOT)/bin/$(TARG)
{{if .LocalDeps}}
# gb: local dependencies{{if $BuildDirPkg=.BuildDirPkg}}
{{range .LocalDeps}}$(TARG): $(GBROOT)/{{$BuildDirPkg}}/{{.}}.a

{{end}}{{end}}{{end}}`))

var MakePkgTemplate = template.Must(template.New("MakePkg").Parse(
	`# Makefile generated by gb: http://go-gb.googlecode.com
# gb provides configuration-free building and distributing

include $(GOROOT)/src/Make.inc

TARG={{.Target}}
GOFILES=\
{{range .GoFiles}}	{{.}}\
{{end}}{{if .AsmObjs}}

OFiles=\
{{range .AsmObjs}}	{{.}}\
{{end}}
{{end}}{{if .CGoFiles}}

CGOFILES=\
{{range .CGoFiles}}	{{.}}\
{{end}}
{{end}}{{if .CObjs}}

CGO_OFILES=\
{{range .CObjs}}	{{.}}\
{{end}}
{{end}}
# gb: this is the local install
GBROOT={{.GBROOT}}

# gb: compile/link against local install
GCIMPORTS+= -I $(GBROOT)/{{.BuildDirPkg}}
LDIMPORTS+= -L $(GBROOT)/{{.BuildDirPkg}}

# gb: compile/link against GOPATH entries
GOPATHSEP=:
ifeq ($(GOHOSTOS),windows)
GOPATHSEP=;
endif
GCIMPORTS+=-I $(subst $(GOPATHSEP),/pkg/$(GOOS)_$(GOARCH) -I , $(GOPATH))/pkg/$(GOOS)_$(GOARCH)
LDIMPORTS+=-L $(subst $(GOPATHSEP),/pkg/$(GOOS)_$(GOARCH) -L , $(GOPATH))/pkg/$(GOOS)_$(GOARCH)
{{if .CopyLocal}}
# gb: copy to local install
$(GBROOT)/{{.BuildDirPkg}}/$(TARG).a: {{.BuildDirPkg}}/$(TARG).a
	mkdir -p $(dir $@); cp -f $< $@
{{end}}
package: $(GBROOT)/{{.BuildDirPkg}}/$(TARG).a

include $(GOROOT)/src/Make.pkg
{{if .LocalDeps}}
# gb: local dependencies{{if $BuildDirPkg=.BuildDirPkg}}
{{range .LocalDeps}}{{$BuildDirPkg}}/$(TARG).a: $(GBROOT)/{{$BuildDirPkg}}/{{.}}.a
{{end}}{{end}}{{end}}`))
