package e

// #include <stdlib.h>
// #include "e4.h"
import "C"

func Atoi(s string) (i int) {
	cs := C.CString(s)
	i = int(C.bar(C.atoi(cs)))
	return
}