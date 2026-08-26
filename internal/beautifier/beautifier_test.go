// internal/beautifier/beautifier_test.go
// Purpose: Unit tests for the Bee source code beautifier engine.
// Responsibility: Verifies that 2-space block indentation, implicit multiplication auto-fix,
//                 and trailing comment alignment are formatted correctly.

package beautifier

import (
	"strings"
	"testing"
)

func TestBeautifierIndentation(t *testing.T) {
	input := `rule main:
new x := 10;
if x > 0 then:
print "positive";
else:
print "non-positive";
done;
return;`

	expected := `rule main:
  new x := 10;
  if x > 0 then:
    print "positive";
  else:
    print "non-positive";
  done;
return;`

	b := New()
	result, err := b.FormatSource(input)
	if err != nil {
		t.Fatalf("FormatSource failed: %v", err)
	}

	if strings.TrimSpace(result) != strings.TrimSpace(expected) {
		t.Errorf("Format mismatch.\nExpected:\n%s\nGot:\n%s", expected, result)
	}
}

func TestImplicitMultiplicationAutoFix(t *testing.T) {
	input := `rule main:
new y := 2(a + b);
return;`

	expected := `rule main:
  new y := 2 * (a + b);
return;`

	b := New()
	result, err := b.FormatSource(input)
	if err != nil {
		t.Fatalf("FormatSource failed: %v", err)
	}

	if strings.TrimSpace(result) != strings.TrimSpace(expected) {
		t.Errorf("Auto-fix mismatch.\nExpected:\n%s\nGot:\n%s", expected, result)
	}
}

func TestCommentAlignment(t *testing.T) {
	input := `rule main:
new x := 10; -- x value
new long_var_name := 20; -- long variable
return;`

	b := New()
	result, err := b.FormatSource(input)
	if err != nil {
		t.Fatalf("FormatSource failed: %v", err)
	}

	if !strings.Contains(result, "-- x value") || !strings.Contains(result, "-- long variable") {
		t.Errorf("Comment alignment failed. Result:\n%s", result)
	}
}
