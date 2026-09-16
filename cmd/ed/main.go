// bee-ed: a minimal, fast, dependency-free file-maintenance CLI for AI agents
// and humans. It applies and inspects diffs, performs safe unique edits, and
// appends reviewable chunks to build new files incrementally.
package main

import (
	"fmt"
	"os"
)

const usage = `bee-ed - file maintenance for AI agents and humans

USAGE:
  bee-ed apply <patch> <file> [--dry-run]
      Apply a unified-diff <patch> to <file>. Atomic; refuses on mismatch.
  bee-ed edit <file> <old> <new> [--dry-run]
      Replace a UNIQUE occurrence of <old> with <new>. Fails on 0 or >1 matches.
  	bee-ed append <file> [chunk|@chunk.txt] [--new]
  	    Append a chunk (arg, @file, or stdin) to <file>. --new requires creation.
    bee-ed balance <file>
        Validate HTML/Markdown tag balance (void, self-closing, comments, code).
    bee-ed sed <pattern> <replacement> <glob...> [--dry-run]
        Regex find-and-replace across a file tree (RE2, parallel worker pool).
    bee-ed apply --help | edit --help | append --help | balance --help | sed --help

FLAGS:
  --dry-run   Show what would change without writing.
  --help, -h  Show help; per-command with <cmd> --help.

All commands verify before writing and produce a concise diff summary.
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	cmd, rest := os.Args[1], os.Args[2:]
	var err error
	switch cmd {
	case "apply":
		err = runApply(rest)
	case "edit":
		err = runEdit(rest)
	case "append":
		err = runAppend(rest)
	case "balance":
		err = runBalance(rest)
	case "sed":
		err = runSed(rest)
	case "--help", "-h", "help":
		fmt.Fprint(os.Stdout, usage)
		return
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n%s", cmd, usage)
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "bee-ed:", err)
		os.Exit(1)
	}
}
