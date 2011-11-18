package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"bufio"
	"bytes"
	"strings"
)

type Config map[string]string

func (cfg Config) Workspace() (dir string, set bool) {
	dir, set = cfg["workspace"]
	return
}

func (cfg Config) Target() (target string, set bool) {
	target, set = cfg["target"]
	return
}

func (cfg Config) AlwaysMakefile() (alwaysMakefile, set bool) {
	amstr, set := cfg["makefile"]
	amstr = strings.ToLower(amstr)
	alwaysMakefile = amstr == "true"
	return
}

func oneLiner(key, path string, cfg Config) {

	val, err := ReadOneLine(path)
	
	if err == nil && val != "" {
		cfg[key] = val
	}

	return
}

var knownKeys = map[string]bool {
	"target": true,
	"workspace": true,
	"makefile": true,
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
