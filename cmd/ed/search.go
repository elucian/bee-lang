package main

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"
	"sync"
)

const searchUsage = `bee-ed search - parallel regex search across a file tree

USAGE:
  bee-ed search <pattern> <glob> [<glob>...] [--count] [--name-only]

<pattern> is a Go RE2 regular expression; it may be "@file" / "@-" to read
the pattern from a file or stdin.

Masks follow the sed rules:
  - A bare glob with no path separator matches the file's BASE NAME at any
    depth (e.g. "*.bee" hits every .bee file under the tree).
  - A glob containing "/" is matched against the full relative path
    (e.g. "spec/*.md").
  - An existing directory (no meta chars) recursively includes all files.

Files are searched concurrently in a worker pool (GOMAXPROCS workers).
Matching lines are reported grep-style as "path:line: text"; files
containing a NUL byte are treated as binary and skipped.

FLAGS:
  --count      Print "path: N" (matching lines per file) instead of lines.
  --name-only  Print only the paths with at least one match.
  --help, -h   Show this help.
`

// searchMatch is one matching line (1-based line number, untrimmed text).
type searchMatch struct {
	line int
	text string
}

// searchResult summarizes one file's outcome.
type searchResult struct {
	path    string
	matches []searchMatch
	binary  bool
	err     error
}

func runSearch(args []string) error {
	var positional []string
	pattern := ""
	count := false
	nameOnly := false

	for _, a := range args {
		switch a {
		case "--count":
			count = true
		case "--name-only":
			nameOnly = true
		case "--help", "-h":
			fmt.Fprint(os.Stdout, searchUsage)
			return nil
		default:
			if strings.HasPrefix(a, "-") {
				return fmt.Errorf("unknown flag %q\n\n%s", a, searchUsage)
			}
			positional = append(positional, a)
		}
	}
	if len(positional) < 2 {
		return fmt.Errorf("usage: bee-ed search <pattern> <glob...>\n\n%s", searchUsage)
	}
	pattern = resolveArg(positional[0])
	masks := positional[1:]

	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid pattern %q: %w", pattern, err)
	}

	jobs, err := collectJobs(masks)
	if err != nil {
		return err
	}
	if len(jobs) == 0 {
		fmt.Fprintf(os.Stdout, "search %q: no files matched %q\n", pattern, masks)
		return nil
	}

	results := runSearchPool(jobs, re)

	totalFiles, totalLines, binaries := 0, 0, 0
	var anyErr error
	for _, r := range results {
		if r.err != nil {
			if anyErr == nil {
				anyErr = r.err
			}
			fmt.Fprintf(os.Stderr, "search %s: %v\n", r.path, r.err)
			continue
		}
		if r.binary {
			binaries++
			continue
		}
		if len(r.matches) == 0 {
			continue
		}
		totalFiles++
		totalLines += len(r.matches)
		switch {
		case nameOnly:
			fmt.Fprintf(os.Stdout, "%s\n", r.path)
		case count:
			fmt.Fprintf(os.Stdout, "%s: %d\n", r.path, len(r.matches))
		default:
			for _, m := range r.matches {
				fmt.Fprintf(os.Stdout, "%s:%d: %s\n", r.path, m.line, m.text)
			}
		}
	}
	fmt.Fprintf(os.Stdout, "search done: %d matching line(s) in %d/%d file(s), %d binary skipped\n",
		totalLines, totalFiles, len(results), binaries)
	return anyErr
}

// runSearchPool inspects jobs in a worker pool and returns results in input
// order. Search is read-only, so no write coordination is needed.
func runSearchPool(jobs []sedJob, re *regexp.Regexp) []searchResult {
	results := make([]searchResult, len(jobs))
	if len(jobs) == 0 {
		return results
	}
	workers := runtime.GOMAXPROCS(0)
	if workers > len(jobs) {
		workers = len(jobs)
	}
	if workers < 1 {
		workers = 1
	}
	jobCh := make(chan int)
	var wg sync.WaitGroup
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go func() {
			defer wg.Done()
			for idx := range jobCh {
				results[idx] = searchOne(jobs[idx], re)
			}
		}()
	}
	for i := range jobs {
		jobCh <- i
	}
	close(jobCh)
	wg.Wait()
	return results
}

// searchOne finds every matching line in one file. A file containing a NUL
// byte is treated as binary and skipped (the tool targets text sources, and
// this keeps object files out of reports).
func searchOne(job sedJob, re *regexp.Regexp) searchResult {
	out := searchResult{path: job.path}
	data, err := os.ReadFile(job.path)
	if err != nil {
		out.err = err
		return out
	}
	if bytes.IndexByte(data, 0) >= 0 {
		out.binary = true
		return out
	}
	for i, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSuffix(line, "\r")
		if re.MatchString(line) {
			out.matches = append(out.matches, searchMatch{line: i + 1, text: line})
		}
	}
	return out
}
