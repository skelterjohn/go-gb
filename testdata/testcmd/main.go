package main

import (
	"testpkg"
	"testcgo"
)

func main() {
	testpkg.Foo()
	testcgo.Foo()
}
