// cmd/bee/main.go
// Purpose: CLI entry point for the Bee Programming Language compiler.
// Responsibility: Handles argument parsing (-c, -b, -e), invokes lexer, parser, and evaluator.
package main

import (
	"bee/internal/beautifier"
	"bee/internal/evaluator"
	"bee/internal/lexer"
	"bee/internal/parser"
	"bee/internal/token"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: bee [-c|-b|-e] [-d] <source.bee>")
		os.Exit(1)
	}

	mode := "-c"
	debug := false
	filePath := ""

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "-c" || arg == "-b" || arg == "-e" || arg == "--compile" || arg == "--beautify" || arg == "--execute" {
			mode = arg
		} else if arg == "-d" || arg == "--debug" {
			debug = true
		} else {
			filePath = arg
		}
	}

	if filePath == "" {
		fmt.Fprintln(os.Stderr, "Error: No source file specified")
		os.Exit(1)
	}

	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", filePath, err)
		os.Exit(1)
	}
	input := string(contentBytes)

	if debug {
		lDebug := lexer.New(input)
		for {
			tok := lDebug.NextToken()
			if tok.Type == token.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "TOKEN: Type=%s, Literal=%q\n", tok.Type, tok.Literal)
		}
	}

	if mode == "-b" || mode == "--beautify" {
		b := beautifier.New()
		formatted, err := b.FormatSource(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Beautify error: %v\n", err)
			os.Exit(1)
		}
		print(formatted)
		return
	}

	// Lex and Parse
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if program == nil {
		fmt.Fprintln(os.Stderr, "Compilation failed: Syntax errors encountered")
		os.Exit(1)
	}

	// Check if any errors occurred during parsing (or basic validation)
	// For compilation check (-c) or evaluation (-e)
	if mode == "-e" {
		eval := evaluator.New()
		eval.Eval(program)
	} else {
		// Compile check: if parser returned nil or failed, exit with 1
		if program == nil {
			fmt.Fprintln(os.Stderr, "Compilation failed: Syntax error")
			os.Exit(1)
		}
	}
}
