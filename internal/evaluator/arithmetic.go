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
		if literal == "³" || literal == "3" {
			rightVal = 3
		} else if literal == "²" || literal == "2" {
			rightVal = 2
		} else if literal == "⁴" || literal == "4" {
			rightVal = 4
		} else if literal == "⁵" || literal == "5" {
			rightVal = 5
		} else if literal == "⁶" || literal == "6" {
			rightVal = 6
		} else if literal == "⁷" || literal == "7" {
			rightVal = 7
		} else if literal == "⁸" || literal == "8" {
			rightVal = 8
		} else if literal == "⁹" || literal == "9" {
			rightVal = 9
		} else if literal != "" && literal != "^" {
			rightVal = parseSuperscriptInt(literal)
		}
		res := 1
		for i := 0; i < rightVal; i++ {
			res *= left
		}
		return res
	case "√":
		deg := left
		if deg == 0 {
			deg = 2
		}
		return int(math.Round(math.Pow(float64(right), 1.0/float64(deg))))
	default:
		if literal != "" && (strings.HasSuffix(literal, "√") || strings.Contains(literal, "√")) {
			deg := 2
			if strings.HasPrefix(literal, "²") {
				deg = 2
			} else if strings.HasPrefix(literal, "³") {
				deg = 3
			} else if strings.HasPrefix(literal, "⁴") {
				deg = 4
			} else {
				orderStr := strings.TrimSuffix(literal, "√")
				if orderStr != "" {
					deg = parseSuperscriptInt(orderStr)
				}
			}
			return int(math.Round(math.Pow(float64(right), 1.0/float64(deg))))
		}
		return 0
	}
}
