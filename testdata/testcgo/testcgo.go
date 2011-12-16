package testcgo

import (
	"unsafe"
)

/*
#include <stdio.h>
#include <stdlib.h>

void myprint(char* s) {
	fprintf(stdout, "%s", s);
}
*/
import "C"

func Foo() {
	var cs *C.char = C.CString("Hello, world!\n")
	C.myprint(cs)
	C.free(unsafe.Pointer(cs))
}
