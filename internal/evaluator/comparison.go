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
	if tokenType == token.NEQ_UNICODE || literal == "≠" || literal == "!=" || strings.Contains(literal, "≠") {
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
	case "<=":
		if left <= right {
			return 1
		}
		return 0
	case ">=":
		if left >= right {
			return 1
		}
		return 0
	}
	return 0
}
