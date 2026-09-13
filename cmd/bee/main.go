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
	p.SetDebug(isDebugging)
	program := p.ParseProgram()
	parseErrors := p.Errors()
	parseWarnings := p.Warnings()

	// Surface lexer deprecation warnings (E0010) emitted during tokenization.
	// Per Decision 3 / Decision 7 — soft warnings that become hard E0009 once
	// Phase 7 audit task 7.2 completes.
	for _, w := range l.Warnings() {
		fmt.Fprintln(os.Stderr, w)
	}

	if len(parseErrors) > 0 {
		for _, e := range parseErrors {
			fmt.Fprintln(os.Stderr, e)
		}
		fmt.Fprintf(os.Stderr, "Parser reported %d error(s).\n", len(parseErrors))
		os.Exit(1)
	}

	// Surface parser warnings (W0901 soft stubs, E0011 deprecated symbols)
	// to stderr but do NOT halt the build. Hardening audit task 7.2 will
	// promote these to E0009. See issues/14-parser-silent-token-drop.md
	// for the W0901/E0009 contract boundary.
	for _, w := range parseWarnings {
		fmt.Fprintln(os.Stderr, w)
	}

	if *executeFlag || *executeLong {
		eval := evaluator.New()
		eval.SetDebug(isDebugging)
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
		fmt.Println("Syntax OK")
		return
	}

	if isDebugging {
		fmt.Println("Syntax OK")
	}
}
