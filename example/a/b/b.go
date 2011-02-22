//target:different/a
package b

import . "c"

func BFoo() {
	println("BFoo")
	CFoo()
}