// cmd/bee/main.go
// Purpose: Main CLI entry point for the Bee compiler toolchain.
// Responsibility: Parses command-line flags, dispatches lexer/parser/beautifier/evaluator,
//                 and reports execution diagnostics.
// Core Architectural Strategy: Zero-dependency Go CLI driver using standard library flag package.

package main

import (
	"bee/internal/beautifier"
	"bee/internal/evaluator"
	"bee/internal/lexer"
	"bee/internal/parser"
	"flag"
	"fmt"
	"os"
)

func main() {
	// Flags
	version := flag.Bool("version", false, "print version")
	v := flag.Bool("v", false, "print version")
	execute := flag.Bool("e", false, "execute code using in-memory VM")
	compile := flag.Bool("c", false, "compile only")
	debug := flag.Bool("d", false, "debug lexer (print token stream)")
	beautify := flag.Bool("b", false, "beautify source code in-place")
	beautifyLong := flag.Bool("beautify", false, "beautify source code in-place")
	dryRun := flag.Bool("dry-run", false, "preview beautification changes without modifying file")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Bee Compiler v0.1.0\n\nUsage: bee [flags] <file.bee>\n\nFlags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *version || *v {
		fmt.Println("Bee Compiler v0.1.0")
		return
	}

	if flag.NArg() < 1 {
		flag.Usage()
		return
	}

	filePath := flag.Arg(0)
	inputBytes, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", filePath, err)
		os.Exit(1)
	}

	input := string(inputBytes)

	// Beautifier Flag Handling (-b / --beautify)
	if *beautify || *beautifyLong {
		b := beautifier.New()
		formatted, err := b.FormatSource(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error beautifying %s: %v\n", filePath, err)
			os.Exit(1)
		}

		if *dryRun {
			fmt.Print(formatted)
			return
		}

		if formatted != input {
			err = os.WriteFile(filePath, []byte(formatted), 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error writing beautified file %s: %v\n", filePath, err)
				os.Exit(1)
			}
			fmt.Printf("Beautified: %s (in-place modification)\n", filePath)
		} else {
			fmt.Printf("Unchanged: %s is already correctly formatted.\n", filePath)
		}
		return
	}

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
