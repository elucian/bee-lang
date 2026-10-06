package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestSplitLines(t *testing.T) {
	tests := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{"a", []string{"a"}},
		{"a\n", []string{"a"}},
		{"a\nb", []string{"a", "b"}},
		{"a\nb\n", []string{"a", "b"}},
		{"a\r\nb\r\n", []string{"a", "b"}},
		{"a\rb", []string{"a", "b"}},
	}
	for _, tc := range tests {
		got := splitLines(tc.in)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("splitLines(%q) = %#v, want %#v", tc.in, got, tc.want)
		}
	}
}

func TestSplitHunkBody(t *testing.T) {
	body := []string{" ctx", "-old", "+new", " ctx"}
	oldBlock, newBlock := splitHunkBody(body)
	wantOld := []string{"ctx", "old", "ctx"}
	wantNew := []string{"ctx", "new", "ctx"}
	if !reflect.DeepEqual(oldBlock, wantOld) {
		t.Errorf("oldBlock = %#v, want %#v", oldBlock, wantOld)
	}
	if !reflect.DeepEqual(newBlock, wantNew) {
		t.Errorf("newBlock = %#v, want %#v", newBlock, wantNew)
	}
}

func TestParseHunkHeader(t *testing.T) {
	cases := []string{
		"@@ -1,3 +2,4 @@ func Foo()",
		"@@ -1 +2 @@",
		"@@ -10,20 +12,30 @@",
	}
	wants := []struct{ old, new int }{
		{1, 2},
		{1, 2},
		{10, 12},
	}
	for i, c := range cases {
		old, new, err := parseHunkHeader(c)
		if err != nil {
			t.Fatalf("parseHunkHeader(%q) error: %v", c, err)
		}
		if old != wants[i].old || new != wants[i].new {
			t.Errorf("parseHunkHeader(%q) = %d,%d want %d,%d", c, old, new, wants[i].old, wants[i].new)
		}
	}
}

func TestParseUnifiedDiff_singleHunk(t *testing.T) {
	text := `--- a/foo.txt
+++ b/foo.txt
@@ -1,3 +1,3 @@
 line1
-line2
+line2b
 line3
`
	hunks, err := parseUnifiedDiff(text)
	if err != nil {
		t.Fatal(err)
	}
	if len(hunks) != 1 {
		t.Fatalf("got %d hunks, want 1", len(hunks))
	}
	h := hunks[0]
	if h.oldStart != 1 || h.newStart != 1 {
		t.Errorf("start = %d,%d want 1,1", h.oldStart, h.newStart)
	}
	ob, nb := splitHunkBody(h.body)
	if !reflect.DeepEqual(ob, []string{"line1", "line2", "line3"}) {
		t.Errorf("old block = %#v", ob)
	}
	if !reflect.DeepEqual(nb, []string{"line1", "line2b", "line3"}) {
		t.Errorf("new block = %#v", nb)
	}
}

func TestApplyHunks_singleChange(t *testing.T) {
	orig := "a\nb\nc"
	hunks, _ := parseUnifiedDiff(`--- a/f
+++ b/f
@@ -1,3 +1,3 @@
 a
-b
+b2
 c
`)
	res, n, err := applyHunks(orig, hunks)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("applied = %d, want 1", n)
	}
	if res != "a\nb2\nc" {
		t.Errorf("result = %q, want %q", res, "a\nb2\nc")
	}
}

func TestApplyHunks_multipleBottomUp(t *testing.T) {
	// Two hunks whose original offsets would collide if applied top-down.
	orig := "l1\nl2\nl3\nl4\nl5"
	hunks, _ := parseUnifiedDiff(`--- a/f
+++ b/f
@@ -1,2 +1,2 @@
 l1
-l2
+X
@@ -3,2 +3,2 @@
 l3
-l4
+Y
`)
	res, n, err := applyHunks(orig, hunks)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Errorf("applied = %d, want 2", n)
	}
	want := "l1\nX\nl3\nY\nl5"
	if res != want {
		t.Errorf("result = %q, want %q", res, want)
	}
}

func TestApplyHunks_shiftedOffsets(t *testing.T) {
	// Insert at top shifts subsequent line numbers if applied in order; the
	// bottom-up applier must keep both hunks valid.
	orig := "a\nb\nc"
	hunks, _ := parseUnifiedDiff(`--- a/f
+++ b/f
@@ -2,1 +2,1 @@
-b
+B
@@ -3,1 +4,1 @@
-c
+C
`)
	res, n, err := applyHunks(orig, hunks)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Errorf("applied = %d, want 2", n)
	}
	want := "a\nB\nC"
	if res != want {
		t.Errorf("result = %q, want %q", res, want)
	}
}

func TestApplyHunks_mismatch(t *testing.T) {
	orig := "a\nX\nc"
	hunks, _ := parseUnifiedDiff(`--- a/f
+++ b/f
@@ -1,3 +1,3 @@
 a
-b
+b2
 c
`)
	res, n, err := applyHunks(orig, hunks)
	if err == nil {
		t.Fatalf("expected error, got result %q n=%d", res, n)
	}
	if !strings.Contains(err.Error(), "mismatch") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestApplyHunks_outOfRange(t *testing.T) {
	orig := "only one line"
	hunks, _ := parseUnifiedDiff(`--- a/f
+++ b/f
@@ -10,2 +10,2 @@
 no
 such
`)
	if _, _, err := applyHunks(orig, hunks); err == nil {
		t.Fatal("expected out-of-range error")
	}
}

func TestApplyHunks_trailingWhitespaceTolerance(t *testing.T) {
	// Trailing whitespace differences should be tolerated by linesEqualFold.
	orig := "a\nb   \nc"
	hunks, _ := parseUnifiedDiff(`--- a/f
+++ b/f
@@ -1,3 +1,3 @@
 a
-b
+B
 c
`)
	res, _, err := applyHunks(orig, hunks)
	if err != nil {
		t.Fatalf("expected tolerance of trailing whitespace, got %v", err)
	}
	if res != "a\nB\nc" {
		t.Errorf("result = %q, want %q", res, "a\nB\nc")
	}
}

func TestHunkTargetStripsABPrefix(t *testing.T) {
	patch := `--- a/dir/file.go
+++ b/dir/file.go
@@ -1,1 +1,1 @@
-x
+x
`
	hunks, err := parseUnifiedDiff(patch)
	if err != nil {
		t.Fatal(err)
	}
	if got := hunkTarget(hunks); got != "dir/file.go" {
		t.Errorf("hunkTarget = %q, want %q", got, "dir/file.go")
	}
	if hunkTarget(hunks) == "" {
		t.Errorf("hunkTarget should not be empty after capture")
	}
}
