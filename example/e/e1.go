package e

// #include <stdlib.h>
import "C"

func Atoi(s string) (i int) {
	cs := C.CString(s)
	i = int(C.atoi(cs))
	return
}