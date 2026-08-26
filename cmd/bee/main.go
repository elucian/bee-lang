package main

import (
	"bee/internal/evaluator"
	"bee/internal/lexer"
	"bee/internal/parser"
	"flag"
	"fmt"
	"os"
)

func main() {
	version := flag.Bool("version", false, "print version")
	v := flag.Bool("v", false, "print version")
	help := flag.Bool("help", false, "print help")
	h := flag.Bool("h", false, "print help")
	execute := flag.Bool("e", false, "execute code")
	compile := flag.Bool("c", false, "compile only")
	debug := flag.Bool("d", false, "debug lexer")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Bee Compiler v0.1.0\n\nUsage: bee [flags] <file.bee>\n\nFlags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *version || *v {
		fmt.Println("Bee Compiler v0.1.0")
		return
	}
	if *help || *h {
		flag.Usage()
		return
	}

	if flag.NArg() < 1 {
		flag.Usage()
		return
	}

	input, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	l := lexer.New(string(input))

	if *debug {
		fmt.Fprintf(os.Stderr, "--- Lexer Debug Output ---\n")
		for tok := l.NextToken(); tok.Type != "EOF"; tok = l.NextToken() {
			fmt.Fprintf(os.Stderr, "%+v\n", tok)
		}
		return
	}

	p := parser.New(l)
	program := p.ParseProgram()

	if *execute {
		eval := evaluator.New()
		eval.Eval(program)
	} else if *compile {
		fmt.Println("Syntax OK: Program parsed successfully.")
	} else {
		flag.Usage()
	}
}
