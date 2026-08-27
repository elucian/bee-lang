// internal/parser/parser_test.go
package parser

import (
	"bee/internal/lexer"
	"testing"
)

func TestParser_Comprehensive(t *testing.T) {
	input := `
	rule testFunc(x Z) => (res Z):
		new a := 10;
		if x > 0 then:
			print "positive";
		else:
			print "non-positive";
		done;
		match x:
			when 1 do
				print "one";
			other:
				print "other";
		done;
		trial:
		try:
			print "trying";
		final
			print "done trying";
		done;
	return a;
	`

	l := lexer.New(input)
	p := New(l)
	prog := p.ParseProgram()

	if prog == nil || len(prog.Statements) == 0 {
		t.Fatalf("Failed to parse program statements")
	}
}
