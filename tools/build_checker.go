package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	fmt.Println("Running 'go build ./...' to check for compilation errors...")

	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = "c:/Users/rain/Documents/GO/Simple C"

	output, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Printf("Build failed with the following errors:\n%s\n", string(output))
		os.Exit(1)
	}

	fmt.Println("Build successful!")
}
