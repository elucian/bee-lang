package main

import (
	"testing"
)

// TestBalanceIgnoresCodeSpanTags guards against the masked-order regression:
// tags documented inside markdown inline code / fences must be blanked BEFORE
// raw-text HTML scanning, otherwise a documented <pre> inside a `code` span is
// mistaken for a real raw-text element that swallows the rest of the file.
func TestBalanceIgnoresCodeSpanTags(t *testing.T) {
	md := "Uses `<pre><code><script>` in a list item."
	problems, open := checkBalance(md)
	if len(problems) != 0 || len(open) != 0 {
		t.Fatalf("got problems=%v open=%v, want none", problems, open)
	}
}

// TestBalanceIgnoresFencedSample guards the fenced-block case: HTML sample code
// inside a ``` fence must not count toward balance.
func TestBalanceIgnoresFencedSample(t *testing.T) {
	md := "doc:\n```html\n<table><tr><td>ok</td></tr></table>\n```\nafter"
	problems, open := checkBalance(md)
	if len(problems) != 0 || len(open) != 0 {
		t.Fatalf("got problems=%v open=%v, want none", problems, open)
	}
}

// TestBalanceDetectsRealUnbalance ensures the mask never hides genuine errors:
// an actually unbalanced element must still be reported.
func TestBalanceDetectsRealUnbalance(t *testing.T) {
	md := "<div>hello" // missing </div>
	problems, open := checkBalance(md)
	if len(problems) != 0 || len(open) == 0 {
		t.Fatalf("got problems=%v open=%v, want an unclosed element", problems, open)
	}
	if open[len(open)-1] != "div" {
		t.Errorf("unclosed = %v, want last element 'div'", open)
	}
}
