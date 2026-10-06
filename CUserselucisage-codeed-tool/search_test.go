package main

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func TestSearchOne_findsMatchesWithLineNumbers(t *testing.T) {
	p := filepath.Join(t.TempDir(), "a.txt")
	if err := os.WriteFile(p, []byte("alpha\nbeta gamma\nnone\nGAMMA\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	res := searchOne(sedJob{path: p}, regexp.MustCompile(`(?i)gamma`))
	if res.err != nil {
		t.Fatal(res.err)
	}
	if len(res.matches) != 2 {
		t.Fatalf("got %d matches, want 2: %+v", len(res.matches), res.matches)
	}
	if res.matches[0].line != 2 || res.matches[0].text != "beta gamma" {
		t.Errorf("match0 = %+v", res.matches[0])
	}
	if res.matches[1].line != 4 || res.matches[1].text != "GAMMA" {
		t.Errorf("match1 = %+v", res.matches[1])
	}
}

func TestSearchOne_noMatch(t *testing.T) {
	p := filepath.Join(t.TempDir(), "n.txt")
	if err := os.WriteFile(p, []byte("nothing here\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	res := searchOne(sedJob{path: p}, regexp.MustCompile("absent"))
	if res.err != nil {
		t.Fatal(res.err)
	}
	if len(res.matches) != 0 {
		t.Errorf("expected no matches, got %+v", res.matches)
	}
}

func TestSearchOne_skipsBinary(t *testing.T) {
	p := filepath.Join(t.TempDir(), "b.bin")
	if err := os.WriteFile(p, []byte{'f', 0, 'o'}, 0o644); err != nil {
		t.Fatal(err)
	}
	res := searchOne(sedJob{path: p}, regexp.MustCompile("o"))
	if !res.binary {
		t.Error("expected file to be detected as binary")
	}
	if len(res.matches) != 0 {
		t.Errorf("binary file should report no matches, got %+v", res.matches)
	}
}

func TestRunSearchPool_preservesInputOrder(t *testing.T) {
	dir := t.TempDir()
	var jobs []sedJob
	for _, name := range []string{"a", "b", "c", "d", "e", "f", "g", "h"} {
		p := filepath.Join(dir, name+".txt")
		if err := os.WriteFile(p, []byte("hit\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		jobs = append(jobs, sedJob{path: p})
	}
	results := runSearchPool(jobs, regexp.MustCompile("hit"))
	for i, r := range results {
		if r.err != nil {
			t.Fatalf("job %d: %v", i, r.err)
		}
		if len(r.matches) != 1 {
			t.Fatalf("job %d matches=%+v", i, r.matches)
		}
		if r.path != jobs[i].path {
			t.Errorf("result %d path=%q, want %q", i, r.path, jobs[i].path)
		}
	}
}
