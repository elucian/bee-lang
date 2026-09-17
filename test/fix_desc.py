# test/fix_desc.py - Bulk @DESC: cleaner for the Bee suite, routed through bee-ed.
#
# Purpose (test/readme.md §4): every `-- @DESC:` block must describe *what the
# test does* and never restate the test's own source. This script enforces that
# across every level in bulk, fixing two kinds of redundancy:
#
#   1. Redundant test-id prefix. Headers like `-- @DESC: T0139 Decision 11: ...`
#      repeat the CASE id that the level README table already shows. The leading
#      test-id token is dropped so the description starts with its subject.
#
#   2. Embedded source from the body. Continuation lines that quote the test's
#      own statements (e.g. `` `new (a, b) : (1, 2) ∈ Z;` ``) duplicate the body.
#      Such lines are reworded to pure prose (a curated mapping, since each is
#      a bespoke sentence) - never mangled by a blind regex.
#
# Metadata comment lines (`@AI`, `@EXPECT`, `@DISABLED`, `@NEGATIVE`,
# `@ENABLED`, `@FROZEN`) are NEVER rewritten: they are intentional documentation
# and may legitimately quote code.
#
# All edits are applied through `bee-ed` (native CLI, built on demand from the
# repo root) - direct hand file-writes are never used. `bee-ed apply` validates
# its unified diff against the working file and refuses to write on any
# mismatch, so the bulk pass is atomic and self-checking per file.
#
# Usage:
#   python test/fix_desc.py            # dry-run: report what WOULD change
#   python test/fix_desc.py --verbose  # dry-run + print rewritten header blocks
#   python test/fix_desc.py --write    # apply changes in place (via bee-ed apply)
import os
import re
import glob
import difflib
import tempfile
import argparse
import subprocess
import sys

# Ensure Python's stdout/stderr can emit Unicode (≠, ≤, ∈, …) on Windows.
try:
    sys.stdout.reconfigure(encoding="utf-8")
    sys.stderr.reconfigure(encoding="utf-8")
except (AttributeError, ValueError):
    pass

TEST_ROOT = os.path.dirname(os.path.abspath(__file__))
REPO_ROOT = os.path.dirname(TEST_ROOT)
BEE_ED = os.path.join(REPO_ROOT, "bin", "bee-ed")

# A leading test-id token right after an `@DESC:` marker, e.g. `T0139 ` / `T0139:`.
LEADING_TEST_ID = re.compile(r"^(--\s*@DESC:)\s*T\d{3,4}\b[\s:]*\s*")

# Metadata tags that must never be rewritten.
META_TAG = re.compile(r"^@[A-Z]+")

# Curated rewrite of continuation lines that embed the test's own source, keyed
# by the exact comment body (everything after `-- `). Rewriting by hand keeps
# the prose intact and descriptive instead of leaving a dangling fragment.
# --- level1 / Decision 13 ---
CONTINUATION_REWRITES = {
    "`(1..5)(1)` discretises the closed range into a 5-element sequence; `[3]` returns the 3rd element (1-based) = 3.":
        "A stepped postfix range yields a 5-element closed sequence; rank-3 indexing gives the third element (1-based) = 3.",
    "`(1>..5)(1)` discretises the left-exclusive range (1,5] to four elements {2,3,4,5}.":
        "Stepping a left-exclusive postfix range (1,5] yields four elements {2,3,4,5}.",
    "`(1>..<5)(1)` discretises the fully-exclusive range (1,5) to three elements {2,3,4}.":
        "Stepping a fully-exclusive postfix range (1,5) yields three elements {2,3,4}.",
    # --- level1 / Decision 11 ---
    "`new (a, b) : (1, 2) ∈ Z;` — N identifiers (N=2), M expressions (M=2), N = M,":
        "Parallel colon-initialisation binds N=2 identifiers to M=2 expressions (N = M),",
    "all sharing the trailing `∈ Z` type qualifier.":
        "all sharing the single trailing type qualifier.",
    # --- level2 / Decision 14 ---
    "Cycle terminates with `done main;`.":
        "Cycle closes with its done terminator after the then epilogue.",
    "D14: `while … do` body closes with `done;` (mandatory terminator).":
        "D14: the while-conditioned body closes with the mandatory done terminator.",
    "`for i ∈ (1..9)(2)` materialises 1, 3, 5, 7, 9 (step 2, 1-based).":
        "Iterates 1, 3, 5, 7, 9 over a step-2 domain (1-based).",
    "`for ∀ i ∈ (1..4)` iterates the inclusive endpoint domain 1..4.":
        "Iterates the inclusive endpoint domain 1..4.",
}


def ensure_bee_ed():
    """Build the bee-ed CLI on demand so edits go through it."""
    if not os.path.exists(BEE_ED):
        r = subprocess.run(["go", "build", "-o", os.path.relpath(BEE_ED, REPO_ROOT),
                            "./cmd/ed/"], cwd=REPO_ROOT, capture_output=True, text=True)
        if r.returncode != 0:
            print(f"error: bee-ed build failed:\n{r.stderr}", file=sys.stderr)
            sys.exit(2)


def process_file(path):
    """Clean the @DESC header block. Returns (changed, new_content)."""
    with open(path, "r", encoding="utf-8") as f:
        content = f.read()
    lines = content.splitlines()

    new_lines = list(lines)
    seen_desc = False
    changed = False

    for idx, raw in enumerate(lines):
        stripped = raw.strip()
        if not stripped.startswith("--"):
            # Header comment block ends at the first non-comment line.
            if not stripped:
                continue
            break

        if "@DESC" in stripped:
            seen_desc = True
            new_raw = LEADING_TEST_ID.sub(r"\1 ", raw, count=1)
            if new_raw != raw:
                new_lines[idx] = new_raw
                changed = True
            continue

        if not seen_desc:
            continue  # comment before @DESC (@DISABLED/@ENABLED etc.) - untouched

        body = re.sub(r"^--\s*", "", stripped)
        if META_TAG.match(body):
            continue  # @AI/@EXPECT/@DISABLED ... - never rewritten

        rewritten = CONTINUATION_REWRITES.get(body)
        if rewritten is not None and rewritten != body:
            indent = re.match(r"^(\s*)", raw).group(1)
            new_lines[idx] = f"{indent}-- {rewritten}"
            changed = True

    if not changed:
        return False, content
    return True, "\n".join(new_lines) + "\n"


def apply_via_bee_ed(rel_path, new_content):
    """Apply the change to rel_path (REPO_ROOT-relative) via `bee-ed apply`."""
    abs_path = os.path.join(REPO_ROOT, rel_path)
    with open(abs_path, "r", encoding="utf-8") as f:
        old_content = f.read()
    if old_content == new_content:
        return True

    diff = "\n".join(difflib.unified_diff(
        old_content.splitlines(),
        new_content.splitlines(),
        fromfile=f"a/{rel_path}", tofile=f"b/{rel_path}", lineterm=""))
    if not diff.strip():
        return True

    with tempfile.NamedTemporaryFile("w", suffix=".patch", delete=False,
                                     encoding="utf-8") as pf:
        pf.write(diff)
        patch_path = pf.name
    try:
        r = subprocess.run([BEE_ED, "apply", patch_path, rel_path],
                           cwd=REPO_ROOT, capture_output=True, text=True)
        if r.returncode != 0:
            print(f"error applying {rel_path}:\n{r.stderr}", file=sys.stderr)
            return False
        return True
    finally:
        os.unlink(patch_path)


def main():
    ap = argparse.ArgumentParser(description="Strip redundancy from @DESC: blocks (via bee-ed)")
    ap.add_argument("--write", action="store_true",
                    help="apply changes in place (default: dry-run)")
    ap.add_argument("--verbose", action="store_true",
                    help="print the rewritten header block per file")
    args = ap.parse_args()

    ensure_bee_ed()

    files = sorted(glob.glob(os.path.join(TEST_ROOT, "level*", "*.bee")))
    changed = []
    for path in files:
        if os.path.basename(path) == "README.md":
            continue
        rel = os.path.relpath(path, REPO_ROOT).replace(os.sep, "/")
        # Echo each test we inspect (no silent bulk pass).
        ok, new_content = process_file(path)
        if not ok:
            print(f"== CHECK {rel}: no change")
            continue
        changed.append((rel, new_content))
        print(f"== CHECK {rel}: would reword")
        if args.verbose:
            with open(path, "r", encoding="utf-8") as f:
                for i, ln in enumerate(f.readlines()[:6], 1):
                    print(f"   {i}: {ln.rstrip()}")
            print()

    touched = len(changed)
    if not args.write:
        print(f"would reword {touched}/{len(files)} bee test files "
              + "(rerun with --write to apply via bee-ed)")
        sys.exit(0)

    ok_writes = 0
    for rel, new_content in changed:
        if apply_via_bee_ed(rel, new_content):
            ok_writes += 1
            print(f"== APPLY {rel}: applied via bee-ed")
        else:
            print(f"== APPLY {rel}: FAILED")
    print(f"wrote {ok_writes}/{touched} bee test files via bee-ed")


if __name__ == "__main__":
    main()
