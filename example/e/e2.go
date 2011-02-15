package e

// #include <stdlib.h>
import "C"

func Atof(s string) (i float64) {
	cs := C.CString(s)
	i = float64(C.atoi(cs))
	return
}