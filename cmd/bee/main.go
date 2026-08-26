package main

import (
	"bee/internal/lexer"
	"bee/internal/token"
	"flag"
	"fmt"
	"os"
)

func main() {
	version := flag.Bool("version", false, "print version")
	v := flag.Bool("v", false, "print version")
	debug := flag.Bool("d", false, "debug lexer")
	flag.Parse()

	if *version || *v {
		fmt.Println("Bee Compiler v0.1.0")
		return
	}

	input, _ := os.ReadFile(flag.Arg(0))
	l := lexer.New(string(input))

	if *debug {
		for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
			fmt.Printf("%+v\n", tok)
		}
		return
	}
}
