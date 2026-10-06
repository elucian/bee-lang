package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

const applyUsage = `bee-ed apply - apply a unified-diff patch

USAGE:
  bee-ed apply <patch> [<file>] [--dry-run]

<patch> may be a file path or "@-" for stdin. If <file> is omitted, the
target is taken from the <patch> header ("--- a/<file>" / "+++ b/<file>").
Hunks are matched against the working file with whitespace tolerance and
applied bottom-up so earlier offsets are unaffected. Refuses to write on
any hunk mismatch.

FLAGS:
  --dry-run  Validate and report hunks without writing.
`

type hunk struct {
	oldStart int // 1-based in original file
	newStart int
	oldCount int
	newCount int
	target   string // "+++ b/<file>" path captured from the diff header
	// body lines with a leading line-type char (' ' context, '-' del, '+' add)
	body []string
}

func runApply(args []string) error {
	var patchSrc, file string
	dryRun := false
	for _, a := range args {
		if strings.HasPrefix(a, "-") {
			switch a {
			case "--help", "-h":
				fmt.Fprint(os.Stdout, applyUsage)
				return nil
			case "--dry-run":
				dryRun = true
			default:
				return fmt.Errorf("unknown flag %q\n\n%s", a, applyUsage)
			}
			continue
		}
		if patchSrc == "" {
			patchSrc = a
		} else if file == "" {
			file = a
		} else {
			return fmt.Errorf("too many arguments (got %q)", a)
		}
	}
	if patchSrc == "" {
		return fmt.Errorf("missing <patch>\n\n" + applyUsage)
	}

	patchText, err := readPatch(patchSrc)
	if err != nil {
		return err
	}
	hunks, err := parseUnifiedDiff(patchText)
	if err != nil {
		return err
	}
	if len(hunks) == 0 {
		return fmt.Errorf("no hunks found in patch")
	}
	if file == "" {
		file = hunkTarget(hunks)
		if file == "" {
			return fmt.Errorf("cannot infer target file from patch header; pass <file> explicitly")
		}
	}

	orig, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	result, nApplied, err := applyHunks(string(orig), hunks)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "apply %s: %d/%d hunks applied\n", file, nApplied, len(hunks))
	if dryRun {
		fmt.Fprintf(os.Stdout, "--- result (%d bytes) below; not written ---\n", len(result))
		fmt.Fprintln(os.Stdout, result)
		return nil
	}
	temp := file + ".bee-ed.tmp"
	if err := os.WriteFile(temp, []byte(result), 0o644); err != nil {
		return fmt.Errorf("write temp: %w", err)
	}
	defer os.Remove(temp)
	if err := os.Rename(temp, file); err != nil {
		return fmt.Errorf("rename onto %q: %w", file, err)
	}
	return nil
}

func readPatch(src string) (string, error) {
	if src == "@-" {
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	b, err := os.ReadFile(src)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// hunkTarget returns the "+++ b/<file>" target stripped of its a/ b/ prefix.
// The target is captured from the diff header during parsing, so when the
// caller omits <file> we can infer it from the patch itself.
func hunkTarget(hunks []hunk) string {
	for _, h := range hunks {
		if h.target != "" {
			return h.target
		}
	}
	return ""
}

// applyHunks applies hunks to orig and returns the new content plus the count
// of hunks successfully applied. Each hunk's old block is validated against
// the ORIGINAL file (so offsets are stable), then all are rebuilt bottom-up.
func applyHunks(orig string, hunks []hunk) (string, int, error) {
	origLines := splitLines(orig)

	// Validate every hunk against the original first.
	applied := 0
	for _, h := range hunks {
		oldBlock, _ := splitHunkBody(h.body)
		idx := h.oldStart - 1
		if idx < 0 || idx+len(oldBlock) > len(origLines) {
			return "", applied, fmt.Errorf("hunk out of range at line %d", h.oldStart)
		}
		if !linesEqualFold(origLines[idx:idx+len(oldBlock)], oldBlock) {
			return "", applied, fmt.Errorf("hunk mismatch at line %d", h.oldStart)
		}
		applied++
	}

	// Apply bottom-up (descending start) on a mutable copy. A change at a
	// higher index never shifts a lower insertion point, so each hunk's
	// original start stays valid against the partially-built result.
	type edit struct {
		idx int
		old []string
		new []string
	}
	var edits []edit
	for _, h := range hunks {
		oldBlock, newBlock := splitHunkBody(h.body)
		edits = append(edits, edit{idx: h.oldStart - 1, old: oldBlock, new: newBlock})
	}
	sort.Slice(edits, func(i, j int) bool { return edits[i].idx > edits[j].idx })

	lines := append([]string(nil), origLines...)
	for _, e := range edits {
		// Copy the tail BEFORE mutating lines: lines[:idx] shares the same
		// backing array as lines[idx+len(old):], so appending e.new in place
		// would clobber the region that `after` still points at.
		after := append([]string(nil), lines[e.idx+len(e.old):]...)
		lines = append(lines[:e.idx], e.new...)
		lines = append(lines, after...)
	}

	return strings.Join(lines, "\n"), applied, nil
}

// splitLines splits on \n, preserving the trailing semantics (no trailing newline element).
func splitLines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	if s == "" {
		return nil
	}
	parts := strings.Split(s, "\n")
	if len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return parts
}

// linesEqualFold compares two line slices, trimming trailing whitespace and
// treating sub-patch "\ No newline" markers as absent.
func linesEqualFold(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if strings.TrimRight(a[i], " \t") != strings.TrimRight(b[i], " \t") {
			return false
		}
	}
	return true
}

// splitHunkBody partitions body lines into old-block and new-block line lists
// (each element is the raw content after the leading type char).
func splitHunkBody(body []string) (oldBlock, newBlock []string) {
	for _, l := range body {
		if len(l) == 0 {
			continue
		}
		switch l[0] {
		case '-':
			oldBlock = append(oldBlock, l[1:])
		case '+':
			newBlock = append(newBlock, l[1:])
		case ' ':
			oldBlock = append(oldBlock, l[1:])
			newBlock = append(newBlock, l[1:])
		}
	}
	return oldBlock, newBlock
}

// parseUnifiedDiff parses a unified diff into hunks.
func parseUnifiedDiff(text string) ([]hunk, error) {
	sc := bufio.NewScanner(strings.NewReader(text))
	var hunks []hunk
	var cur *hunk
	target := ""
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "+++ ") && target == "" {
			// Capture the new-file path, e.g. "+++ b/dir/file.go". Set it on
			// current and future hunks so hunkTarget can infer the target.
			target = stripABPrefix(strings.TrimSpace(line[4:]))
			if cur != nil {
				cur.target = target
			}
			continue
		}
		if strings.HasPrefix(line, "@@") {
			hs, hn, err := parseHunkHeader(line)
			if err != nil {
				return nil, err
			}
			if cur != nil {
				hunks = append(hunks, *cur)
			}
			cur = &hunk{oldStart: hs, newStart: hn, target: target}
			continue
		}
		if cur == nil {
			continue // file header lines (---/+++) and other metadata
		}
		if len(line) == 0 {
			// empty context line
			cur.body = append(cur.body, " ")
			continue
		}
		c := line[0]
		if c == ' ' || c == '-' || c == '+' || c == '\\' {
			if c == '\\' {
				// "\ No newline at end of file" - ignore for matching.
				continue
			}
			cur.body = append(cur.body, line)
		}
	}
	if sc.Err() != nil {
		return nil, sc.Err()
	}
	if cur != nil {
		hunks = append(hunks, *cur)
	}
	// Compute counts if not present.
	for i := range hunks {
		if hunks[i].oldCount == 0 {
			ob, _ := splitHunkBody(hunks[i].body)
			hunks[i].oldCount = len(ob)
		}
		if hunks[i].newCount == 0 {
			_, nb := splitHunkBody(hunks[i].body)
			hunks[i].newCount = len(nb)
		}
	}
	return hunks, nil
}

// stripABPrefix removes the "a/" or "b/" prefix that git's unified diff
// headers use (“--- a/path“ / “+++ b/path“). Bare paths without a known
// prefix are returned unchanged.
func stripABPrefix(p string) string {
	switch {
	case len(p) > 2 && (strings.HasPrefix(p, "a/") || strings.HasPrefix(p, "b/")):
		return p[2:]
	default:
		return p
	}
}

// parseHunkHeader parses "@@ -l[,c] +l[,c] @@ extra".
func parseHunkHeader(line string) (oldStart, newStart int, err error) {
	inner := line
	if i := strings.Index(inner, "@@ "); i >= 0 {
		inner = inner[i+3:]
	}
	if i := strings.LastIndex(inner, " @@"); i >= 0 {
		inner = inner[:i]
	}
	parts := strings.Fields(inner)
	if len(parts) < 2 {
		return 0, 0, fmt.Errorf("bad hunk header %q", line)
	}
	return parseCount(parts[0]), parseCount(parts[1]), nil
}

// parseCount parses "-l,c" or "-l" returning l (the start line).
func parseCount(s string) int {
	s = strings.TrimLeft(s, "-+")
	comma := strings.Index(s, ",")
	if comma >= 0 {
		s = s[:comma]
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}
