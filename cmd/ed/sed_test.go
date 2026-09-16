package main

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func TestMatchPath_normalizesSeparators(t *testing.T) {
	ok, err := matchPath("tutorial/*.html", "tutorial/operators.html")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Error("expected tutorial/*.html to match tutorial/operators.html")
	}
	ok, err = matchPath("tutorial/*.html", "tutorial/sub/operators.html")
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("tutorial/*.html must not cross a subdirectory boundary (path.Match semantics)")
	}
}

func TestPathRoot(t *testing.T) {
	cases := map[string]string{
		"*.bee":           "",
		"tutorial/*.html": "tutorial",
		"a/b/*.go":        "a/b",
	}
	for mask, want := range cases {
		if got := pathRoot(mask); got != want {
			t.Errorf("pathRoot(%q) = %q, want %q", mask, got, want)
		}
	}
}

func TestSedOne_replacesAndWrites(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(p, []byte("foo bar foo\nhello foo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	re := regexp.MustCompile("foo")
	res := sedOne(sedJob{path: p}, re, "XYZ", false)
	if res.err != nil {
		t.Fatal(res.err)
	}
	if !res.matched || res.replaced != 3 {
		t.Errorf("matched=%v replaced=%d, want true/3", res.matched, res.replaced)
	}
	got, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "XYZ bar XYZ\nhello XYZ\n" {
		t.Errorf("file after write = %q", string(got))
	}
}

func TestSedOne_groupCaptureExpansion(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "c.txt")
	if err := os.WriteFile(p, []byte("deep foo deep\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	re := regexp.MustCompile(`(\w+) foo`)
	res := sedOne(sedJob{path: p}, re, "$1-XX", false)
	if res.err != nil {
		t.Fatal(res.err)
	}
	got, _ := os.ReadFile(p)
	if string(got) != "deep-XX deep\n" {
		t.Errorf("file after group expansion = %q", string(got))
	}
}

func TestSedOne_noMatchLeavesFileUntouched(t *testing.T) {
	p := filepath.Join(t.TempDir(), "n.txt")
	if err := os.WriteFile(p, []byte("nothing here\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	re := regexp.MustCompile("foo")
	res := sedOne(sedJob{path: p}, re, "X", false)
	if res.err != nil {
		t.Fatal(res.err)
	}
	if res.matched {
		t.Error("expected no match")
	}
}

func TestRunPool_processesInParallelAndWrites(t *testing.T) {
	dir := t.TempDir()
	var jobs []sedJob
	want := []string{"a", "b", "c", "d", "e", "f"}
	for _, name := range want {
		p := filepath.Join(dir, name+".txt")
		if err := os.WriteFile(p, []byte("foo\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		jobs = append(jobs, sedJob{path: p, base: filepath.Base(p)})
	}
	re := regexp.MustCompile("foo")
	results := runPool(jobs, re, "bar", false)
	for i, r := range results {
		if r.err != nil {
			t.Fatalf("job %d: %v", i, r.err)
		}
		if !r.matched || r.replaced != 1 {
			t.Errorf("job %s matched=%v replaced=%d", r.path, r.matched, r.replaced)
		}
		got, _ := os.ReadFile(r.path)
		if string(got) != "bar\n" {
			t.Errorf("job %s content = %q", r.path, string(got))
		}
	}
}

func TestRunPool_dryRunDoesNotWrite(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "d.txt")
	if err := os.WriteFile(p, []byte("foo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	results := runPool([]sedJob{{path: p}}, regexp.MustCompile("foo"), "bar", true)
	if results[0].err != nil {
		t.Fatal(results[0].err)
	}
	got, _ := os.ReadFile(p)
	if string(got) != "foo\n" {
		t.Errorf("dry-run should not modify file, got %q", string(got))
	}
}
