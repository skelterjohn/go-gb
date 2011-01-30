package a

import (
	"fmt"
	"gomatrix.googlecode.com/hg/matrix"
)

func AFoo() {
	println("AFoo")
	fmt.Printf("%v\n", matrix.Eye(2))
}
