package evaluator

func evalLogical(op string, left, right int) int {
	switch op {
	case "and", "∧":
		if left != 0 && right != 0 {
			return 1
		}
		return 0
	case "or", "∨":
		if left != 0 || right != 0 {
			return 1
		}
		return 0
	case "xor", "⊕":
		if (left != 0) != (right != 0) {
			return 1
		}
		return 0
	case "¬":
		if right == 0 {
			return 1
		}
		return 0
	}
	return 0
}
