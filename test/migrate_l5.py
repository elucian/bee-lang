# test/migrate_l5.py - Level 5 test migration driver to the boxed-doc format.
#
# References T0501.bee, the approved example of the new autonomous/assertion-only
# documentation style. Converts the remaining O/C level5 cases (T0502..T0517) to
# that shape:
#
#   +----...----            header block comment carrying the wrapped @DESC
#   -- @DESC: <line1>
#   --        <line2 ...>   (only when the description exceeds ~50 chars)
#   ----...----+
#   rule main: ...
#   return;
#
#   +----...----            footer block comment carrying the spec / rationale
#   <prose ...>
#   ----...----+
#   -- @AI: No - assertion-only (self-verifying)
#   -- @DISABLED: <reason>
#
# Every write is routed through `bee-ed apply` (unified diff) so no file is ever
# hand-edited; `bee-ed` refuses to write on a hunk mismatch. The block-comment
# boxes use the left-border-only `+---` / `---+` pattern the compiler lexer
# balances via its `+-`/`-+` delimiters (see config/AGENTS.md, internal/lexer).
#
# Usage:
#   python test/migrate_l5.py            # dry-run: print diffs, change nothing
#   python test/migrate_l5.py --apply    # write the diffs via bee-ed apply
import os
import re
import difflib
import tempfile
import argparse
import subprocess
import sys

TEST_ROOT = os.path.dirname(os.path.abspath(__file__))
REPO_ROOT = os.path.dirname(TEST_ROOT)
LVL5 = os.path.join(TEST_ROOT, "level5")
BEE_ED = os.path.join(REPO_ROOT, "bin", "bee-ed")

try:
    sys.stdout.reconfigure(encoding="utf-8")
    sys.stderr.reconfigure(encoding="utf-8")
except (AttributeError, ValueError):
    pass

DESC_WIDTH = 50   # keep @DESC lines short; wrap keeping whole words
PROSE_WIDTH = 60  # tidy footer box prose


def wrap(text, width):
    """Wrap `text` at `width` on whole words."""
    lines, cur = [], ""
    for w in text.split():
        t = (cur + " " + w).strip()
        if len(t) <= width:
            cur = t
        else:
            if cur:
                lines.append(cur)
            cur = w
    if cur:
        lines.append(cur)
    return lines or [text]


def header_box(desc):
    """Build the header block comment: `+---` ... `@DESC` ... `---+`."""
    parts = wrap(desc, DESC_WIDTH)
    lines = ["-- @DESC: " + parts[0]]
    pad = "-- " + (" " * len("@DESC: "))
    for p in parts[1:]:
        lines.append(pad + p)
    w = max(len(l) for l in lines) + 2
    top = "+" + "-" * (w - 1)
    bottom = "-" * (w - 1) + "+"
    return [top] + lines + [bottom]


def footer_box(prose):
    """Build the footer block comment from plain prose lines (no `--` prefix)."""
    lines = wrap(" ".join(prose), PROSE_WIDTH)
    w = max(len(l) for l in lines) + 2
    top = "+" + "-" * (w - 1)
    bottom = "-" * (w - 1) + "+"
    return [top] + lines + [bottom]


def find_return(lines):
    """Index of the LAST `return;` line (the main rule's return)."""
    idx = None
    for i, l in enumerate(lines):
        if l.strip().startswith("return;"):
            idx = i
    return idx


def split_file(path):
    """Parse a legacy T05XX.bee into its parts.

    Returns (desc, body, prose, tags) where `tags` are the normalized footer
    metadata lines (`-- @AI: ...`, `-- @DISABLED: ...` etc., in original order,
    with @DISABLED continuations folded in).
    """
    with open(path, "r", encoding="utf-8", newline="") as f:
        lines = f.read().split("\n")
    if lines and lines[-1] == "":
        lines.pop()

    body_start = None
    for i, l in enumerate(lines):
        s = l.strip()
        if s and not s.startswith("--"):
            body_start = i
            break
    if body_start is None:
        raise SystemExit(f"{path}: no body found")

    header = lines[:body_start]
    ret = find_return(lines)
    if ret is None:
        raise SystemExit(f"{path}: no return; found")
    body = lines[body_start:ret + 1]
    footer = lines[ret + 1:]

    desc = None
    for l in header:
        if "@DESC:" in l:
            desc = l.split("@DESC:")[1].strip()
            break
    if not desc:
        desc = "TBD"

    prose, tags = [], []
    in_disabled = False
    for l in footer:
        s = l.strip()
        if s.startswith("--"):
            inner = s[2:].strip()
            if not inner:
                continue
            if inner.startswith("@"):
                tags.append("-- " + inner)
                in_disabled = "@DISABLED" in inner
            elif in_disabled:
                tags.append("-- " + inner)  # continuation of the @DISABLED reason
            else:
                prose.append(inner)
    return desc, body, prose, tags


def build_new(desc, body, prose, tags):
    out = []
    out += header_box(desc)
    out += body
    out.append("")
    if prose:
        out += footer_box(prose)
    out += tags
    return "\n".join(out) + "\n"


def read_text(path):
    with open(path, "r", encoding="utf-8", newline="") as f:
        return f.read()


def apply_via_bee_ed(rel_path, new_content):
    abs_path = os.path.join(REPO_ROOT, rel_path)
    old_content = read_text(abs_path)
    if old_content == new_content:
        return True
    diff = "\n".join(difflib.unified_diff(
        old_content.splitlines(), new_content.splitlines(),
        fromfile="a/" + rel_path, tofile="b/" + rel_path, lineterm=""))
    with tempfile.NamedTemporaryFile("w", suffix=".patch", delete=False,
                                     encoding="utf-8") as pf:
        pf.write(diff)
        patch = pf.name
    try:
        r = subprocess.run([BEE_ED, "apply", patch, rel_path],
                           cwd=REPO_ROOT, capture_output=True, text=True)
        if r.returncode != 0:
            return False  # caller prints stderr
        return True
    finally:
        os.unlink(patch)


def files_to_migrate():
    fs = sorted(os.listdir(LVL5))
    return [os.path.join(LVL5, f) for f in fs
            if f.endswith(".bee") and f.startswith("T05") and f != "T0501.bee"]


def verify(path):
    """Re-parse a migrated file and assert the harness still reads its tags."""
    import test  # noqa: F401  (importable parent package is `test/`)
    with open(path, "r", encoding="utf-8", newline="") as f:
        lines = f.readlines()
    disabled = any("@DISABLED" in ln for ln in lines)
    desc = next((ln.split("@DESC:")[1].strip() for ln in lines
                 if "@DESC:" in ln), "MISSING")
    body = list(test.strip_comments(lines))
    print(f"    verify {os.path.basename(path)}: desc='{desc}' "
          f"disabled={disabled} body_lines={len(body)}")


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--apply", action="store_true",
                    help="write diffs via bee-ed apply (default: dry-run)")
    args = ap.parse_args()

    if not os.path.exists(BEE_ED):
        r = subprocess.run(["go", "build", "-o", "bin/bee-ed", "./cmd/ed/"],
                           cwd=REPO_ROOT, capture_output=True, text=True)
        if r.returncode != 0:
            sys.exit("bee-ed build failed:\n" + r.stderr)

    changed = []
    for path in files_to_migrate():
        desc, body, prose, tags = split_file(path)
        new_content = build_new(desc, body, prose, tags)
        rel = os.path.relpath(path, REPO_ROOT).replace(os.sep, "/")
        old = read_text(path)
        if old == new_content:
            print(f"== {rel}: unchanged")
            continue
        changed.append((rel, new_content))
        ndiff = len(list(difflib.unified_diff(old.splitlines(), new_content.splitlines())))
        print(f"== {rel}: {ndiff} diff lines")

    print(f"\nwould migrate {len(changed)}/{len(files_to_migrate())} level5 files "
          + "(rerun with --apply to write via bee-ed)")
    if not args.apply:
        return

    ok = 0
    for rel, new_content in changed:
        if apply_via_bee_ed(rel, new_content):
            ok += 1
            print(f"== APPLY {rel}: ok")
        else:
            print(f"== APPLY {rel}: FAILED")
    print(f"wrote {ok}/{len(changed)} files via bee-ed")
    if ok:
        print("verifying migrated files through the harness parser:")
        for rel, _ in changed:
            verify(os.path.join(REPO_ROOT, rel))


if __name__ == "__main__":
    main()
