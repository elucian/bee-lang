# scripts/migrate_repeat_to_next.py
#
# Decision 15 (2026-09-13): `next` is the canonical loop-jump keyword;
# `repeat` is a deprecated synonym that lexes to NEXT with an E0010 warning.
#
# This script migrates `.bee` test sources from the legacy `repeat` spelling
# to the canonical `next` spelling. The replacement is word-boundary aware so
# it rewrites the keyword in code and in @DESC/comment keyword mentions, but
# leaves English prose such as "repeated"/"repeats" untouched.
#
# Usage:
#   python scripts/migrate_repeat_to_next.py            # migrate all test/**/*.bee
#   python scripts/migrate_repeat_to_next.py --dry-run  # report only, write nothing
#   python scripts/migrate_repeat_to_next.py <path...>  # migrate specific files/dirs
#
# Exit code is 0 whether or not changes were made; --dry-run exits 1 when it
# finds files that would change (useful as a CI lint gate).

import os
import re
import sys

# Whole-word `repeat`, not preceded/followed by identifier chars. Bee
# identifiers are alphanumeric + underscore, so \b is sufficient.
REPEAT_WORD = re.compile(r"\brepeat\b")

DEFAULT_ROOTS = ["test"]


def iter_bee_files(roots):
    for root in roots:
        if os.path.isfile(root):
            if root.endswith(".bee"):
                yield root
            continue
        for dirpath, _dirs, files in os.walk(root):
            for name in files:
                if name.endswith(".bee"):
                    yield os.path.join(dirpath, name)


def migrate_file(path, dry_run):
    with open(path, "r", encoding="utf-8") as f:
        original = f.read()

    # Per-line replacement so we can report line numbers.
    changed_lines = []
    out_lines = []
    for lineno, line in enumerate(original.splitlines(keepends=True), 1):
        new_line, n = REPEAT_WORD.subn("next", line)
        if n:
            changed_lines.append((lineno, n))
        out_lines.append(new_line)

    if not changed_lines:
        return 0

    if not dry_run:
        with open(path, "w", encoding="utf-8") as f:
            f.write("".join(out_lines))

    total = sum(n for _ln, n in changed_lines)
    tag = "would update" if dry_run else "updated"
    print(f"{tag}: {path} ({total} replacement(s) on line(s) "
          f"{', '.join(str(ln) for ln, _n in changed_lines)})")
    return total


def main():
    args = [a for a in sys.argv[1:] if a != "--dry-run"]
    dry_run = "--dry-run" in sys.argv
    roots = args if args else DEFAULT_ROOTS

    total_replacements = 0
    files_changed = 0
    for path in iter_bee_files(roots):
        n = migrate_file(path, dry_run)
        if n:
            files_changed += 1
            total_replacements += n

    if files_changed == 0:
        print("No `repeat` occurrences found. Nothing to do.")
        return 0

    verb = "Would migrate" if dry_run else "Migrated"
    print(f"{verb} {total_replacements} occurrence(s) across "
          f"{files_changed} file(s).")
    # Dry-run with pending changes exits non-zero so it can gate CI.
    return 1 if dry_run else 0


if __name__ == "__main__":
    sys.exit(main())
