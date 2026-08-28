package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// The health check is self-contained and validates the build and core functionality.

func main() {
	fmt.Println("--- Starting Bee Compiler Self-Health Check ---")

	// 1. Verify environment
	if _, err := exec.LookPath("go"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: 'go' not found: %v\n", err)
		os.Exit(1)
	}

	// 2. Build compiler
	fmt.Println("-> Verifying compiler build...")
	absPath, _ := os.Getwd()
	beeCompiler := filepath.Join(absPath, "bin", "bee-compiler.exe")
	buildCmd := exec.Command("go", "build", "-o", beeCompiler, "./cmd/bee/main.go")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "Build failed:\n%s\n", out)
		os.Exit(1)
	}
	defer os.Remove(beeCompiler)

	// 3. Smoke test with smoke.bee
	fmt.Println("-> Running integrated smoke test...")
	smokeTest := "test/level0/smoke.bee"

	runCmd := exec.Command(beeCompiler, "-c", smokeTest)
	out, err := runCmd.CombinedOutput()

	outputStr := string(out)
	if !strings.Contains(outputStr, "Syntax OK") {
		fmt.Fprintf(os.Stderr, "Smoke test did not report 'Syntax OK'. Output:\n%s\nError: %v\n", outputStr, err)
		os.Exit(1)
	}

	fmt.Println("--- Health Check PASSED ---")
}
