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
	"bufio"
	"bytes"
	//"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config map[string]string

func (cfg Config) ProtobufPlugin() (plugin string, set bool) {
	plugin, set = cfg["proto"]
	return
}

func (cfg Config) Workspace() (dir string, set bool) {
	dir, set = cfg["workspace"]
	return
}

func (cfg Config) Target() (target string, set bool) {
	target, set = cfg["target"]
	return
}

func (cfg Config) Ignore() (ignore, set bool) {
	ignore, set = cfg.IgnoreAll()
	if ignore && set {
		return
	}
	vstr, set := cfg["ignore"]
	vstr = strings.ToLower(vstr)
	ignore = vstr == "true"
	if t, s := cfg.Target(); s {
		ignore = ignore || t == "-"
	}
	return
}

func (cfg Config) IgnoreAll() (ignoreAll, set bool) {
	vstr, set := cfg["ignoreall"]
	vstr = strings.ToLower(vstr)
	ignoreAll = vstr == "true"
	if t, s := cfg.Target(); s {
		ignoreAll = ignoreAll || t == "--"
	}
	return
}

func (cfg Config) AlwaysMakefile() (alwaysMakefile, set bool) {
	amstr, set := cfg["makefile"]
	amstr = strings.ToLower(amstr)
	alwaysMakefile = amstr == "true"
	return
}

func (cfg Config) GCFlags() (gcflags string, set bool) {
	gcflags, set = cfg["gcflags"]
	return
}

func (cfg Config) Write(dir string) (err error) {
	path := filepath.Join(dir, "gb.cfg")
	var fout *os.File
	fout, err = os.Create(path)
	if err != nil {
		return
	}

	for key, val := range cfg {
		fmt.Fprintf(fout, "%s=%s\n", key, val)
	}

	fout.Close()

	os.Remove(filepath.Join(dir, "target.gb"))
	os.Remove(filepath.Join(dir, "workspace.gb"))

	return
}

func oneLiner(key, path string, cfg Config) {

	val, err := ReadOneLine(path)

	if err == nil && val != "" {
		cfg[key] = val
	}

	return
}

var knownKeys = map[string]bool{
	"proto":     true,
	"target":    true,
	"workspace": true,
	"makefile":  true,
	"ignore":    true,
	"ignoreall": true,
	"gcflags":   true,
}

func ReadConfig(dir string) (cfg Config) {
	cfg = make(map[string]string)

	path := filepath.Join(dir, "gb.cfg")
	fin, existserr := os.Open(path)
	if existserr == nil {

		br := bufio.NewReader(fin)

		for {
			line, isPrefix, brerr := br.ReadLine()
			if brerr != nil {
				break
			}
			if isPrefix {
				ErrLog.Println(errors.New(fmt.Sprintf("config line too long: %s", path)))
				break
			}

			line = bytes.TrimSpace(line)
			if len(line) == 0 {
				continue
			}

			split := bytes.Index(line, []byte("="))
			if split == -1 {
				ErrLog.Println(errors.New(fmt.Sprintf("config line malformed: %s", path)))
				break
			}
			key, val := line[:split], line[split+1:]
			key = bytes.ToLower(bytes.TrimSpace(key))
			val = bytes.TrimSpace(val)
			cfg[string(key)] = string(val)
			if !knownKeys[string(key)] {
				ErrLog.Printf("Unknown key '%s' in config %s", key, path)
			}
		}
	}

	oneLiner("target", filepath.Join(dir, "target.gb"), cfg)
	oneLiner("workspace", filepath.Join(dir, "workspace.gb"), cfg)

	return
}
