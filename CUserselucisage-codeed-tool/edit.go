package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

const editUsage = `bee-ed edit - safe, unique substring replacement

USAGE:
  bee-ed edit <file> <old> <new> [--dry-run]

<old> is replaced with <new> ONLY when it occurs exactly once in <file>. It
is an error to have zero or multiple occurrences (anti-guess guarantee, so
an agent can never silently patch the wrong spot).

Pass "@path" to read <old> or <new> from a file (useful for multi-line and
HTML snippets). Use "@-" to read from stdin.
`

func runEdit(args []string) error {
	var file, old, new string
	dryRun := false

	// Collect positional args so --dry-run may appear anywhere.
	var positional []string
	for _, a := range args {
		switch a {
		case "--dry-run":
			dryRun = true
		case "--help", "-h":
			fmt.Fprint(os.Stdout, editUsage)
			return nil
		default:
			if strings.HasPrefix(a, "-") {
				return fmt.Errorf("unknown flag %q\n\n%s", a, editUsage)
			}
			positional = append(positional, a)
		}
	}
	for i, a := range positional {
		switch i {
		case 0:
			file = a
		case 1:
			old = a
		case 2:
			new = a
		default:
			return fmt.Errorf("too many positional arguments (got %q)", a)
		}
	}
	if file == "" || old == "" {
		return errors.New("usage: bee-ed edit <file> <old> <new>\n\n" + editUsage)
	}

	// Resolve @file references for old/new.
	old = resolveArg(old)
	new = resolveArg(new)

	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	if old == "" {
		return errors.New("<old> resolved to empty; nothing to replace")
	}

	idx := strings.Index(string(content), old)
	if idx < 0 {
		return fmt.Errorf("no exact match for <old> in %q", file)
	}
	if strings.Index(string(content)[idx+len(old):], old) >= 0 {
		return fmt.Errorf("ambiguous edit: <old> occurs more than once in %q; refusing", file)
	}

	updated := string(content)[:idx] + new + string(content)[idx+len(old):]
	fmt.Fprintf(os.Stdout, "edit %s: replaced unique match (%d -> %d bytes)\n",
		file, len(old), len(updated)-len(old)+len(old))
	if dryRun {
		fmt.Fprintf(os.Stdout, "--- result (%d bytes) below; not written ---\n", len(updated))
		fmt.Fprintln(os.Stdout, updated)
		return nil
	}
	temp := file + ".bee-ed.tmp"
	if err := os.WriteFile(temp, []byte(updated), 0o644); err != nil {
		return fmt.Errorf("write temp: %w", err)
	}
	defer os.Remove(temp)
	if err := os.Rename(temp, file); err != nil {
		return fmt.Errorf("rename onto %q: %w", file, err)
	}
	return nil
}

// resolveArg expands "@path" -> file contents and "@-" -> stdin; otherwise
// returns the argument unchanged.
func resolveArg(a string) string {
	if a == "@-" {
		b, err := io.ReadAll(os.Stdin)
		if err == nil {
			return string(b)
		}
		return ""
	}
	if strings.HasPrefix(a, "@") && len(a) > 1 {
		b, err := os.ReadFile(a[1:])
		if err == nil {
			return string(b)
		}
	}
	return a
}
