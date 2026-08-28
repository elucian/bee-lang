package evaluator

import (
	"math"
	"strings"
)

func parseSuperscriptInt(s string) int {
	val := 0
	for _, r := range s {
		digit := -1
		switch r {
		case '⁰':
			digit = 0
		case '¹':
			digit = 1
		case '²':
			digit = 2
		case '³':
			digit = 3
		case '⁴':
			digit = 4
		case '⁵':
			digit = 5
		case '⁶':
			digit = 6
		case '⁷':
			digit = 7
		case '⁸':
			digit = 8
		case '⁹':
			digit = 9
		}
		if digit >= 0 {
			val = val*10 + digit
		}
	}
	if val == 0 {
		return 2
	}
	return val
}

func evalSqrt(literal string, rightVal int) int {
	deg := 2
	if literal != "" {
		orderStr := strings.TrimSuffix(literal, "√")
		if orderStr != "" {
			parsed := parseSuperscriptInt(orderStr)
			if parsed > 0 {
				deg = parsed
			}
		}
	}
	res := math.Round(math.Pow(float64(rightVal), 1.0/float64(deg)))
	return int(res)
}

func evalArithmetic(op string, left, right int, literal string) int {
	switch op {
	case "+":
		return left + right
	case "-":
		return left - right
	case "*":
		return left * right
	case "/":
		if right != 0 {
			return left / right
		}
		return 0
	case "%":
		if right != 0 {
			res := left % right
			if res < 0 {
				if right > 0 {
					res += right
				} else {
					res -= right
				}
			}
			return res
		}
		return 0
	case "^":
		rightVal := right
		if literal == "³" {
			rightVal = 3
		} else if literal == "²" {
			rightVal = 2
		} else if literal == "⁴" {
			rightVal = 4
		} else if literal == "⁵" {
			rightVal = 5
		} else if literal == "⁶" {
			rightVal = 6
		} else if literal == "⁷" {
			rightVal = 7
		} else if literal == "⁸" {
			rightVal = 8
		} else if literal == "⁹" {
			rightVal = 9
		} else if literal != "" && literal != "^" {
			parsed := parseSuperscriptInt(literal)
			if parsed > 0 && literal != "^" && literal != "1" && literal != "2" && literal != "3" && literal != "4" && literal != "5" && literal != "6" && literal != "7" && literal != "8" && literal != "9" {
				rightVal = parsed
			}
		}
		if rightVal == 0 && literal == "0" {
			rightVal = 0
		}
		res := 1
		for i := 0; i < rightVal; i++ {
			res *= left
		}
		return res
	case "√":
		if right != 0 {
			return evalSqrt(literal, right)
		}
		return evalSqrt(literal, left)
	default:
		if literal != "" && (strings.HasSuffix(literal, "√") || strings.Contains(literal, "√")) {
			if left != 0 {
				return evalSqrt(literal, left)
			}
			return evalSqrt(literal, right)
		}
		return 0
	}
}
