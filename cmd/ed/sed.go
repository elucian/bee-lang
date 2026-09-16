package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
)

const sedUsage = `bee-ed sed - regex find-and-replace across a file tree (parallel)

USAGE:
  bee-ed sed <pattern> <replacement> <glob> [<glob>...] [--dry-run]

<pattern> is a Go RE2 regular expression; <replacement> uses standard
regexp expansion ($1, ${name}; literal $ is written as $$). Both may be
"@file" / "@-" to read the value from a file/stdin.

Masks are matched against the walk:
  - A bare glob with no path separator matches the file's BASE NAME at any
    depth (e.g. "*.bee" hits every .bee file under the tree).
  - A glob containing "/" is matched against the full relative path
    (e.g. "tutorial/*.html").
  - An existing directory (no meta chars) recursively includes all files.

Files are processed concurrently in a worker pool (say GOMAXPROCS workers). A
file is rewritten atomically (temp file + rename) only if a match replaced
something, and it refuses to touch files that are not regular files.

FLAGS:
  --dry-run  Report matches without writing any file.
`

// sedJob carries one file to process; the run gatherer and per-file state are
// threaded through the worker pool via the job/result channel pair.
type sedJob struct {
	path string
	base string
}

// sedResult summarizes one file's outcome.
type sedResult struct {
	path     string
	matched  bool
	replaced int
	bytesIn  int64
	bytesOut int64
	err      error
}

func runSed(args []string) error {
	var positional []string
	pattern, replacement := "", ""
	masks := []string{}
	dryRun := false

	for _, a := range args {
		switch a {
		case "--dry-run":
			dryRun = true
		case "--help", "-h":
			fmt.Fprint(os.Stdout, sedUsage)
			return nil
		default:
			if strings.HasPrefix(a, "-") {
				return fmt.Errorf("unknown flag %q\n\n%s", a, sedUsage)
			}
			positional = append(positional, a)
		}
	}
	for i, a := range positional {
		switch i {
		case 0:
			pattern = a
		case 1:
			replacement = a
		default:
			masks = append(masks, a)
		}
	}
	if pattern == "" || len(masks) == 0 {
		return fmt.Errorf("usage: bee-ed sed <pattern> <replacement> <glob...>\n\n" + sedUsage)
	}
	// An empty replacement is allowed and means DELETE every match:
	// ReplaceAllString with "" drops the matched text.
	// Note: consecutive empty replacements can produce adjacent duplicates;
	// batch non-empty patterns first, then delete leftovers.

	pattern = resolveArg(pattern)
	replacement = resolveArg(replacement)

	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid pattern %q: %w", pattern, err)
	}

	jobs, err := collectJobs(masks)
	if err != nil {
		return err
	}
	if len(jobs) == 0 {
		fmt.Fprintf(os.Stdout, "sed %q -> %q: no files matched %q\n", pattern, replacement, masks)
		return nil
	}

	results := runPool(jobs, re, replacement, dryRun)

	// Aggregate + report.
	totalMatched, totalReplaced := 0, 0
	var anyErr error
	for _, r := range results {
		if r.err != nil {
			if anyErr == nil {
				anyErr = r.err
			}
			fmt.Fprintf(os.Stderr, "sed %s: %v\n", r.path, r.err)
			continue
		}
		if r.matched {
			totalMatched++
			totalReplaced += r.replaced
			if dryRun {
				fmt.Fprintf(os.Stdout, "sed %s: would replace %d bytes (%d -> %d)\n",
					r.path, r.replaced, r.bytesIn, r.bytesOut)
			} else {
				fmt.Fprintf(os.Stdout, "sed %s: %d match(es) (%d -> %d bytes)\n",
					r.path, r.replaced, r.bytesIn, r.bytesOut)
			}
		}
	}
	verb := "replaced"
	if dryRun {
		verb = "would replace"
	}
	fmt.Fprintf(os.Stdout, "sed done: %d/%d files %s, %d total match(es)\n",
		totalMatched, len(results), verb, totalReplaced)
	return anyErr
}

// collectJobs walks each mask and returns the ordered list of regular files
// to process. Duplicates are folded away.
func collectJobs(masks []string) ([]sedJob, error) {
	unique := map[string]bool{}
	var jobs []sedJob
	for _, mask := range masks {
		hasSep := strings.ContainsAny(mask, "/\\")
		if hasSep {
			// Full-path mask: walk the tree rooted at the segment before the
			// first meta char and match the full relative path against mask.
			root := pathRoot(mask)
			if root == "" {
				root = "."
			}
			err := filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					return nil
				}
				ok, me := matchPath(mask, p)
				if me != nil {
					return me
				}
				if ok {
					addJob(p, info, &jobs, unique)
				}
				return nil
			})
			if err != nil {
				return nil, fmt.Errorf("walk %q: %w", root, err)
			}
			continue
		}
		// Meta-free mask: if it is an existing directory, include all files
		// recursively; otherwise match the file base name at any depth.
		if fi, err := os.Stat(mask); err == nil && fi.IsDir() {
			err := filepath.Walk(mask, func(p string, info os.FileInfo, werr error) error {
				if werr != nil {
					return werr
				}
				if info == nil || info.IsDir() {
					return nil
				}
				addJob(p, info, &jobs, unique)
				return nil
			})
			if err != nil {
				return nil, fmt.Errorf("walk %q: %w", mask, err)
			}
			continue
		}
		// Name glob: recurse from "." and match against the base name.
		root := "."
		err := filepath.Walk(root, func(p string, info os.FileInfo, werr error) error {
			if werr != nil {
				return werr
			}
			if info == nil || info.IsDir() {
				return nil
			}
			ok, me := filepath.Match(mask, info.Name())
			if me != nil {
				return me
			}
			if ok {
				addJob(p, info, &jobs, unique)
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("walk %q: %w", mask, err)
		}
	}
	return jobs, nil
}

// matchPath normalises path separators to '/' (so masks authored with '/'
// match walk paths that use the OS separator on Windows) before glob-matching.
func matchPath(mask, p string) (bool, error) {
	if os.PathSeparator != '/' {
		mask = strings.ReplaceAll(mask, string(os.PathSeparator), "/")
		p = strings.ReplaceAll(p, string(os.PathSeparator), "/")
	}
	return path.Match(mask, p)
}

func addJob(p string, info os.FileInfo, jobs *[]sedJob, unique map[string]bool) {
	if !info.Mode().IsRegular() {
		return // sockets, pipes, symlinks, dirs
	}
	if unique[p] {
		return
	}
	unique[p] = true
	*jobs = append(*jobs, sedJob{path: p, base: filepath.Base(p)})
}

// runPool processes jobs in a worker pool and returns the results in input
// order. Reads and rewrites are independent per file, so no coordination is
// needed beyond the atomic temp+rename write.
func runPool(jobs []sedJob, re *regexp.Regexp, replacement string, dryRun bool) []sedResult {
	results := make([]sedResult, len(jobs))
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
				results[idx] = sedOne(jobs[idx], re, replacement, dryRun)
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

func sedOne(job sedJob, re *regexp.Regexp, replacement string, dryRun bool) sedResult {
	out := sedResult{path: job.path}
	data, err := os.ReadFile(job.path)
	if err != nil {
		out.err = err
		return out
	}
	out.bytesIn = int64(len(data))
	// ReplaceAllString performs a single pass; a replacement equal to the
	// original match leaves the byte count unchanged (no-op).
	src := string(data)
	if !re.MatchString(src) {
		out.matched = false
		out.bytesOut = out.bytesIn
		return out
	}
	updated := re.ReplaceAllString(src, replacement)
	if updated == src {
		out.matched = true // engine matched but produced no net change
		out.replaced = 0
		out.bytesOut = out.bytesIn
		return out
	}
	out.matched = true
	// Each distinct match site counts as one substitution, even if the
	// expansion is zero-width or reproduces the original text.
	out.replaced = len(re.FindAllStringIndex(src, -1))
	out.bytesOut = int64(len(updated))
	if dryRun {
		return out
	}
	temp := job.path + ".bee-ed.tmp"
	if err := os.WriteFile(temp, []byte(updated), 0o644); err != nil {
		out.err = fmt.Errorf("write temp: %w", err)
		return out
	}
	defer os.Remove(temp)
	if err := os.Rename(temp, job.path); err != nil {
		out.err = fmt.Errorf("rename onto %q: %w", job.path, err)
		return out
	}
	return out
}

// pathRoot returns the leading non-meta directory segment of a glob so a
// full-path mask can seed a bounded walk (e.g. "tutorial/*.html" -> "tutorial").
func pathRoot(mask string) string {
	segs := strings.FieldsFunc(mask, func(r rune) bool {
		return r == '/' || r == '\\'
	})
	var root string
	for _, s := range segs {
		if strings.ContainsAny(s, "*?[") {
			break
		}
		if root == "" {
			root = s
		} else {
			root += "/" + s
		}
	}
	return root
}
