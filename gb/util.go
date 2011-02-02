package main

import (
	"os"
	"strings"
	"fmt"
	"path"
)

func GetAbsolutePath(p string) (absp string, err os.Error) {
	if path.IsAbs(p) {
		absp = p
		return
	}
	wd, err := os.Getwd()
	if p == "." {
		absp = wd
		return
	}
	absp = path.Join(wd, p)
	return
}

func GetRelativePath(parent, child string) (rel string, err os.Error) {
	parent, err = GetAbsolutePath(parent)
	child, err = GetAbsolutePath(child)

	if !strings.HasPrefix(child, parent) {
		err = os.NewError(fmt.Sprintf("'%s' is not in '%s'", child, parent))
	}

	rel = child[len(parent)+1 : len(child)]

	return
}
