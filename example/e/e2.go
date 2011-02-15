package e
// #cgo LDFLAGS: -lm
// #include <math.h>
import "C"

func CSin(x float64) (y float64) {
	y = float64(C.sin(_Ctype_double(x)))
	return
}