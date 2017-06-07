package flock

import (
	"fmt"
	"os"
)

// we need to add Test case instead of Example.
func ExampleFlock() {
	fname := "./test.txt"
	flag := os.O_RDWR | os.O_CREATE
	f, e := os.OpenFile(fname, flag, 0600)
	if e != nil {
		fmt.Printf("Failed to create a file. %v", e)
	}
	defer f.Close()

	// FIXME(Shaohe, Feng): does not find to run a process to flock the file.
	// such as: exec.Command("flock", "-x", fname, "-c", "cat")
	fmt.Println(Flock(f, 0600, true, 0))
	fmt.Println(Funlock(f))
	// Output:
	// <nil>
	// <nil>
}
