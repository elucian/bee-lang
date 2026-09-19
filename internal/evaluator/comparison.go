package evaluator

import (
	"bee/internal/token"
	"strings"
)

func evalComparison(op string, left, right int, tokenType token.Type, literal string) int {
	if tokenType == token.EQ || literal == "=" || literal == "==" {
		if left == right {
			return 1
		}
		return 0
	}
	if tokenType == token.NEQ || tokenType == token.NEQ_UNICODE || literal == "¬" || literal == "≠" || literal == "!=" || literal == "<>" || strings.Contains(literal, "≠") {
		if left != right {
			return 1
		}
		return 0
	}
	switch op {
	case "<":
		if left < right {
			return 1
		}
		return 0
	case ">":
		if left > right {
			return 1
		}
		return 0
	case "<=", "≤":
		if left <= right {
			return 1
		}
		return 0
	case ">=", "≥":
		if left >= right {
			return 1
		}
		return 0
	}
	return 0
}

// approxDefaultEps is the default tolerance for ≈ (spec/05 §4.2): the
// $max_precision default of 1e-5. Integer operands make the default window
// exact-equality; an explicit `± t` modifier widens it inclusively.
const approxDefaultEps = 1e-5

// evalApproxEqual reports 1 when |left − right| ≤ tol (spec/05 §4.2).
func evalApproxEqual(left, right int, tol float64) int {
	d := left - right
	if d < 0 {
		d = -d
	}
	if float64(d) <= tol {
		return 1
	}
	return 0
}
