package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const appendUsage = `bee-ed append - append a reviewable chunk to a file

USAGE:
  bee-ed append <file> [chunk|@chunk_file] [--new] [--dry-run]

The chunk may be passed directly as an argument, read from an @file, or read
from stdin when omitted. The append is atomic and prints a summary.

FLAGS:
  --new      Require that <file> does not already exist (create a fresh file).
  --dry-run  Show what would be appended without writing.
`

func runAppend(args []string) error {
	var file string
	var chunkSrc string
	dryRun := false
	requireNew := false

	for _, a := range args {
		switch {
		case a == "--dry-run":
			dryRun = true
		case a == "--new":
			requireNew = true
		case a == "--help", a == "-h":
			fmt.Fprint(os.Stdout, appendUsage)
			return nil
		case strings.HasPrefix(a, "-"):
			return fmt.Errorf("unknown flag %q\n\n%s", a, appendUsage)
		case file == "":
			file = a
		case chunkSrc == "":
			chunkSrc = a
		default:
			return fmt.Errorf("too many arguments (already have file=%q chunk=%q, got %q)", file, chunkSrc, a)
		}
	}
	if file == "" {
		return errors.New("missing <file>\n\n" + appendUsage)
	}

	chunk := chunkSrc
	if chunkSrc != "" {
		if strings.HasPrefix(chunkSrc, "@") {
			b, err := os.ReadFile(chunkSrc[1:])
			if err != nil {
				return err
			}
			chunk = string(b)
		}
	} else {
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		chunk = string(b)
	}
	if chunk == "" {
		return errors.New("nothing to append (empty chunk)")
	}
	if !strings.HasSuffix(chunk, "\n") {
		chunk += "\n"
	}

	exists, err := pathExists(file)
	if err != nil {
		return err
	}
	if exists && requireNew {
		return fmt.Errorf("refusing --new: %q already exists", file)
	}
	if !exists {
		if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
			return err
		}
	}

	fmt.Fprintf(os.Stdout, "append %s: +%d lines, %dB\n",
		file, strings.Count(chunk, "\n"), len(chunk))
	if dryRun {
		return nil
	}
	f, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.WriteString(chunk); err != nil {
		return err
	}
	return f.Sync()
}

func pathExists(p string) (bool, error) {
	_, err := os.Stat(p)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}
