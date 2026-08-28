package main

import (
	"bee/internal/beautifier"
	"bee/internal/evaluator"
	"bee/internal/lexer"
	"bee/internal/parser"
	"bee/internal/token"
	"flag"
	"fmt"
	"os"
	"runtime"
)

func main() {
	compileFlag := flag.Bool("c", false, "Compile/validate syntax")
	compileLong := flag.Bool("compile", false, "Compile/validate syntax")
	executeFlag := flag.Bool("e", false, "Execute in-memory VM")
	executeLong := flag.Bool("execute", false, "Execute in-memory VM")
	debugFlag := flag.Bool("d", false, "Debug mode (tokens & call stack info)")
	debugLong := flag.Bool("debug", false, "Debug mode (tokens & call stack info)")
	beautifyFlag := flag.Bool("b", false, "Beautify source code")
	beautifyLong := flag.Bool("beautify", false, "Beautify source code")

	flag.Parse()
	args := flag.Args()

	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: No input file specified.")
		os.Exit(1)
	}

	filePath := args[0]
	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", filePath, err)
		os.Exit(1)
	}
	content := string(contentBytes)

	isDebugging := *debugFlag || *debugLong

	if isDebugging {
		_, goFile, goLine, _ := runtime.Caller(0)
		fmt.Fprintf(os.Stderr, "[DEBUG TRACE] Source File: %s | Invoked from Go: %s:%d\n", filePath, goFile, goLine)
		l := lexer.New(content)
		for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
			fmt.Fprintf(os.Stderr, "[TOKEN] Type: %-15s | Literal: %-10s | Pos: %d\n", tok.Type, tok.Literal, tok.Pos)
		}
	}

	if *beautifyFlag || *beautifyLong {
		b := beautifier.New()
		formatted, err := b.FormatSource(content)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[BEAUTIFY ERROR] %v\n", err)
			os.Exit(1)
		}
		fmt.Println(formatted)
		return
	}

	l := lexer.New(content)
	p := parser.New(l)
	program := p.ParseProgram()

	if *executeFlag || *executeLong {
		eval := evaluator.New()
		if isDebugging {
			_, goFile, goLine, _ := runtime.Caller(0)
			fmt.Fprintf(os.Stderr, "[EXECUTION TRACE] Initializing VM Evaluator | Go Origin: %s:%d\n", goFile, goLine)
		}
		eval.Eval(program)
		if isDebugging {
			fmt.Fprintf(os.Stderr, "[EXECUTION TRACE] Program execution completed successfully.\n")
		}
		return
	}

	// TODO: Compile to LLVM IR (Not implemented yet).
	// Redirecting to syntax check.
	if *compileFlag || *compileLong {
		fmt.Fprintf(os.Stderr, "Warning: Compilation to LLVM IR is not implemented yet. Running syntax check.\n")
		if len(p.Errors()) > 0 {
			for _, msg := range p.Errors() {
				fmt.Fprintf(os.Stderr, "Syntax Error: %s\n", msg)
			}
			os.Exit(1)
		}
		fmt.Println("Syntax OK")
		return
	}

	if isDebugging {
		fmt.Println("Syntax OK")
	}
}
