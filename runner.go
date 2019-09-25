package autotest

import "fmt"

// RunTests will run a new set of tests whenever a file changes
func RunTests(filename string) error {
	fmt.Println("Changed file", filename)
	return nil
}
