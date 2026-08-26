// cmd/bee/main.go
// Purpose: Main CLI entry point for the Bee compiler toolchain.
// Responsibility: Parses command-line flags, dispatches lexer/parser/beautifier/evaluator,
//                 and reports execution diagnostics.
// Core Architectural Strategy: Zero-dependency Go CLI driver using standard library flag and filepath packages.

package main

import (
	"bee/internal/beautifier"
	"bee/internal/evaluator"
	"bee/internal/lexer"
	"bee/internal/parser"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const compilerVersion = "Bee Compiler v0.1.0"

func printHelp() {
	helpText := fmt.Sprintf(`%s
A fast, elegant, rule-oriented programming language compiler.

USAGE:
  bee [flags] <file.bee | directory>

COMMANDS & FLAGS:
  -e                 Execute source code using the in-memory AST evaluator VM.
  -c                 Compile/parse only and verify syntax without execution.
  -b, --beautify     Beautify source code in-place (enforce 2-space indentation, align comments, auto-fix).
  --dry-run          Preview beautification changes in stdout without writing to disk.
  -d                 Debug mode: print lexical token stream to stderr.
  -v, --version      Display compiler version information.
  -h, --help         Display this help message and command overview.

EXAMPLES:
  bee hello.bee                     # Parse and check syntax
  bee -e hello.bee                  # Execute code in VM
  bee -b hello.bee                  # Format and beautify file in-place
  bee -b demo/                      # Format all .bee files in directory
  bee -b --dry-run hello.bee        # Preview formatting diff
  bee -d hello.bee                  # Inspect token stream for debugging
`, compilerVersion)
	fmt.Fprintf(os.Stderr, "%s\n", helpText)
}

func beautifyFile(filePath string, b *beautifier.Beautifier, dryRun bool) error {
	inputBytes, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read error: %w", err)
	}

	input := string(inputBytes)
	formatted, err := b.FormatSource(input)
	if err != nil {
		return fmt.Errorf("formatting error: %w", err)
	}

	if dryRun {
		fmt.Printf("--- Dry-Run: %s ---\n", filePath)
		fmt.Print(formatted)
		return nil
	}

	if formatted != input {
		err = os.WriteFile(filePath, []byte(formatted), 0644)
		if err != nil {
			return fmt.Errorf("write error: %w", err)
		}
		fmt.Printf("Beautified: %s (in-place modification)\n", filePath)
	} else {
		fmt.Printf("Unchanged:  %s (already correctly formatted)\n", filePath)
	}
	return nil
}

func main() {
	// Flags
	version := flag.Bool("version", false, "display compiler version")
	v := flag.Bool("v", false, "display compiler version")
	help := flag.Bool("help", false, "display help documentation")
	h := flag.Bool("h", false, "display help documentation")
	execute := flag.Bool("e", false, "execute code using in-memory VM")
	compile := flag.Bool("c", false, "compile only")
	debug := flag.Bool("d", false, "debug lexer (print token stream)")
	beautify := flag.Bool("b", false, "beautify source code in-place")
	beautifyLong := flag.Bool("beautify", false, "beautify source code in-place")
	dryRun := flag.Bool("dry-run", false, "preview beautification changes without modifying file")

	flag.Usage = printHelp
	flag.Parse()

	if *version || *v {
		fmt.Println(compilerVersion)
		return
	}

	if *help || *h {
		printHelp()
		return
	}

	if flag.NArg() < 1 {
		printHelp()
		return
	}

	targetPath := flag.Arg(0)
	if targetPath == "help" {
		printHelp()
		return
	}

	fileInfo, err := os.Stat(targetPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error accessing target '%s': %v\n", targetPath, err)
		os.Exit(1)
	}

	// Beautifier Flag Handling (-b / --beautify)
	if *beautify || *beautifyLong {
		b := beautifier.New()

		if fileInfo.IsDir() {
			fmt.Printf("Beautifying all .bee files in directory: %s\n", targetPath)
			err := filepath.Walk(targetPath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.IsDir() && strings.HasSuffix(info.Name(), ".bee") {
					if err := beautifyFile(path, b, *dryRun); err != nil {
						fmt.Fprintf(os.Stderr, "Error beautifying %s: %v\n", path, err)
					}
				}
				return nil
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "Directory walk error: %v\n", err)
				os.Exit(1)
			}
			return
		}

		if err := beautifyFile(targetPath, b, *dryRun); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Standard Compilation / Execution Pipeline
	inputBytes, err := os.ReadFile(targetPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", targetPath, err)
		os.Exit(1)
	}
	input := string(inputBytes)

	l := lexer.New(input)

	// Debug mode: Print lexer output
	if *debug {
		fmt.Fprintf(os.Stderr, "--- Lexer Debug Output ---\n")
		for tok := l.NextToken(); tok.Type != "EOF"; tok = l.NextToken() {
			fmt.Fprintf(os.Stderr, "%+v\n", tok)
		}
	}

	// Parsing stage: Always run
	p := parser.New(l)
	program := p.ParseProgram()

	// VM Evaluator / Compile-only check
	if *execute {
		eval := evaluator.New()
		eval.Eval(program)
	} else if *compile {
		fmt.Println("Syntax OK: Program parsed successfully.")
	}
}
