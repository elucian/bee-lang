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
	execute := flag.Bool("e", false, "execute code")
	compile := flag.Bool("c", false, "compile only")
	flag.Parse()

	input, _ := os.ReadFile(flag.Arg(0))
	l := lexer.New(string(input))
	p := parser.New(l)
	program := p.ParseProgram()

	if *execute {
		eval := evaluator.New()
		eval.Eval(program)
	} else if *compile {
		fmt.Println("Syntax OK")
	}
}
