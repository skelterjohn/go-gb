//target:g
package g
/* This is a package in a sub-dir where the sub-dir is not a namespace.
   gb thinks this target is called 'f/g', but we call it 'g'.
   The comment just before the 'package g' corrects this.
   We could instead have a file called f/g/target.gb which contains:
g
   In other words, just the full name of the target, all by itself.
*/

func GFoo() {
	println("GFoo")
}
