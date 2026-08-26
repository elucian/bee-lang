package main

import (
	"bee/internal/lexer"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: bee <file.bee>")
		return
	}

	// Basic lexing test
	input, _ := os.ReadFile(os.Args[1])
	l := lexer.New(string(input))

	for tok := l.NextToken(); tok.Type != "EOF"; tok = l.NextToken() {
		fmt.Printf("%+v\n", tok)
	}
}
