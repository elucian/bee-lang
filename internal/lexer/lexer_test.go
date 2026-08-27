// internal/lexer/lexer_test.go
package lexer

import (
	"bee/internal/token"
	"testing"
)

func TestLexer_Comprehensive(t *testing.T) {
	input := `
	rule testFunc(x Z) => (res Z):
		new a := 10..20;
		new b := 1.!10;
		new c := 1!.10;
		new d := 1!!10;
		new e := x :: y;
		new f := x +> chan;
		new g := chan << msg;
		new h := λ(x) => (x * 2);
		new s := "val = #(a + b)";
		new r := ` + "`raw string`" + `;
		<sql>SELECT * FROM users</sql>
		let val := 5 + (: nested (: comment :) comment :) 2;
	return;
	`

	l := New(input)

	for {
		tok := l.NextToken()
		if tok.Type == token.EOF {
			break
		}
		if tok.Type == token.ILLEGAL {
			t.Fatalf("unexpected ILLEGAL token: %q", tok.Literal)
		}
	}
}
