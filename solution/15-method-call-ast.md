# Solution 15 — Method-call AST & Resolution

Approach proposed to close `issues/15-method-call-grant.md`.

## 1. AST
Add to `internal/ast/ast.go`:
```go
type MethodCallExpression struct {
    Token     token.Token
    Receiver  Expression
    Method    string
    Arguments []Expression
}

func (mce *MethodCallExpression) expressionNode() {}
func (mce *MethodCallExpression) TokenLiteral() string { return mce.Token.Literal }
func (mce *MethodCallExpression) String() string {
    args := []string{}
    for _, a := range mce.Arguments { args = append(args, a.String()) }
    return mce.Receiver.String() + "." + mce.Method + "(" + strings.Join(args, ", ") + ")"
}
```

## 2. Lexer
Add `DOT` token type in `internal/token/token.go`. The `.` is ambiguous between member-export prefix (`spec/04 §4.1`) and method receiver (this spec). Resolved by **position-only disambiguation**:
- Top-level grammar: `.beep` after a `member_decl` opener → export prefix.
- Inside an expression position: `.type` after an identifier → method receiver.

Lexer-level: just emit a `DOT` token; grammar disambiguates.

## 3. Parser
In `parseExpression`, after the `left` identifier block, when `peekChar == '.'`:
```go
if p.l.PeekChar() == '.' {
    p.l.NextToken() // consume '.'
    methodTok := p.l.NextToken()
    if methodTok.Type != token.IDENT {
        p.errors = append(p.errors, ...)
        return nil
    }
    var args []Expression
    if p.l.PeekChar() == '(' {
        p.l.NextToken() // consume '('
        if p.l.PeekChar() != ')' {
            args = append(args, p.parseExpression())
            for p.l.PeekChar() == ',' {
                p.l.NextToken()
                args = append(args, p.parseExpression())
            }
        }
        p.l.NextToken() // consume ')'
    }
    left = &MethodCallExpression{Token: methodTok, Receiver: left, Method: methodTok.Literal, Arguments: args}
}
```

## 4. Evaluator
Add type-method dispatch table in `internal/evaluator/evaluator.go`:
```go
var builtinMethods = map[string]map[string]BuiltinMethod{
    "Z": {"type": func(...) Value { return StringValue("Z") }},
    "R": {"type": func(...) Value { return StringValue("R") }},
    "U": {"type": func(...) Value { return StringValue("U") }},
    "B": {"type": func(...) Value { return StringValue("B") }},
}
```

In `Eval`, when given a `*ast.MethodCallExpression`:
1. Evaluate the receiver.
2. Look up the receiver's type tag in `builtinMethods`.
3. If method exists, call it with the args. Otherwise `E0706: MethodNotFound`.

## 5. Verification
- Run `test/debt/T0126-method-call.bee` standalone: `python test/solo.py T0126-method-call`.
- Promote the test back to `test/level1/` (drop `@DEBT` marker, copy to `test/level1/T0126.bee`, update `test/level1/README.md`).
- Run `sh run.sh test level1`: green.
- Run `sh run.sh smoke`: green.

## Existing Precedent
Dotted access is already used at the top-level grammar (`module.member` in `spec/04 §4.2`). The compiler currently handles it via the `member_decl` opener. The new `MethodCallExpression` is the EXPRESSION-side counterpart.
