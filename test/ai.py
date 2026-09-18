# test/ai.py - AI-aware test runner for the Bee suite.
#
# Design goal (config/AGENTS.md §5 + test/readme.md §4): every .bee case must
# pass or fail on its own - never rely on a human (or agent) eyeballing output.
#
# This runner adds one explicit layer for output-producing tests:
#
#   * A test is "AI based" (AI=Yes) when its SOURCE BODY uses `print` (or any
#     stdout side channel), declared in the header with the tag
#     `-- @AI: Yes - <reason>`. Purely assertion-driven tests (only `expect`)
#     are self-verifying and are marked @AI: No.
#   * An AI based test MUST declare its exact expected stdout via the header
#     tag `-- @EXPECT: <exact output>` (see test/readme.md §4). The runner
#     captures the program's stdout, normalizes it, and compares it to the
#     declared expectation. A MISMATCH (or a missing @EXPECT) makes the test
#     FAIL - no human (or agent) reading required.
#   * The runner prints a machine-readable verdict per case
#       [T0001] base=PASS AI=Yes expect="..." actual="..." -> OK
#     so an AI agent driving the pipeline can confirm the printed output
#     matches expectation deterministically.
#
# Usage:
#   python test/ai.py [level]      # run a level (default level0)
#   python test/ai.py T0001        # run a single named case
#   python test/ai.py --nv         # non-verbose (suppress full stdout/stderr)
import os
import re
import sys
import json
import argparse
import datetime
import subprocess

# Ensure Python's stdout/stderr can emit Unicode (≠, ≤, ≥, …) on Windows and
# other terminals whose libc encoding is not UTF-8.
try:
    sys.stdout.reconfigure(encoding="utf-8")
    sys.stderr.reconfigure(encoding="utf-8")
except (AttributeError, ValueError):
    pass

LEVELS = ["level0", "level1", "level2", "level3", "level4", "level5",
          "level6", "level7", "level8", "level20"]
BEE = "./bin/bee.exe"
PRINT_RE = re.compile(r"\bprint\b")


def decode_expect(raw):
    """Decode `\n`/`\t` escapes in an @EXPECT tag so a multi-line expected
    output can be spelled on one comment line (AI-verifiable stdout)."""
    if raw is None:
        return None
    return raw.replace("\\n", "\n").replace("\\t", "\t").strip()


def esc(s):
    """Render a captured/expected string safely on a single machine-readable
    line: real newlines/tabs become \\n / \\t escapes."""
    if s is None:
        return "None"
    return s.replace("\\", "\\\\").replace("\n", "\\n").replace("\t", "\\t")


def parse_header(path):
    """Return (desc, expect, negative, disabled, ai_tag) from a .bee file.

    Design (test/readme.md §4): line 1 carries only `-- @DESC:`; every
    lifecycle/verification directive (@AI/@EXPECT/@NEGATIVE/@DISABLED) lives
    in the trailing FOOTER comments. The whole file is scanned so tags are
    honored regardless of exact position.

    ai_tag is tri-state:
      True  -> footer declares `-- @AI: Yes` (output must be machine-checked)
      False -> footer declares `-- @AI: No` (self-sufficient, authoritative)
      None  -> no @AI tag (fall back to uses_print auto-detect)
    """
    desc, expect, negative, disabled, ai_tag = "TBD", None, False, False, None
    try:
        with open(path, "r", encoding="utf-8") as tf:
            for line in tf:
                if "@DISABLED" in line:
                    disabled = True
                if "@NEGATIVE" in line:
                    negative = True
                if "-- @DESC:" in line and desc == "TBD":
                    # First @DESC (line 1) wins; later occurrences are ignored.
                    desc = line.split("@DESC:")[1].strip()
                if "-- @EXPECT:" in line:
                    expect = decode_expect(line.split("@EXPECT:")[1].strip())
                m = re.search(r"@AI:\s*(Yes|No)", line)
                if m:
                    ai_tag = m.group(1) == "Yes"
    except Exception:
        pass
    return desc, expect, negative, disabled, ai_tag


def ai_used_for(ai_tag, path):
    """An explicit `@AI: No` header is authoritative: a self-sufficient
    assertion-heavy test does not become AI-based just because it contains a
    `print` fallback line (e.g. a negative compile-rejection test whose print
    is dead code). When no @AI tag is present, fall back to print detection.
    """
    if ai_tag is False:
        return False
    if ai_tag is True:
        return True
    return uses_print(path)


def uses_print(path):
    """AI is used iff the SOURCE BODY emits stdout via the `print` statement.

    Header comment lines (`-- @DESC:`, `-- @AI:`, `-- @EXPECT:`, …) never count,
    so a description that merely says "print" cannot mark an assertion-only
    test AI.
    """
    try:
        with open(path, "r", encoding="utf-8") as tf:
            for line in tf:
                if line.lstrip().startswith("--"):
                    continue
                if PRINT_RE.search(line):
                    return True
    except Exception:
        pass
    return False


def run_case(test_name, path, verbose):
    """Run one case. Returns (name, status, ai_used, reason, verdict, expect, actual)."""
    desc, expect, negative, disabled, ai_tag = parse_header(path)
    ai_used = ai_used_for(ai_tag, path)
    base = "PASS"
    reason = ""

    if disabled:
        return test_name, "SKIP", ai_used, "(disabled)", "SKIP", None, None, desc

    res = subprocess.run([BEE, "-e", path], capture_output=True, text=True,
                         encoding="utf-8", errors="replace")

    if negative:
        base = "PASS" if res.returncode != 0 else "FAIL"
    else:
        base = "PASS" if res.returncode == 0 else "FAIL"

    if verbose:
        print(f"--- STDOUT ({path}) ---")
        print(res.stdout)
        print(f"--- STDERR ({path}) ---")
        print(res.stderr)

    actual = res.stdout.strip()
    verdict = "OK"

    if ai_used:
        if expect is None or expect == "":
            status = "FAIL"
            reason = "missing @EXPECT header tag on print-based test"
            verdict = "NO-EXPECT"
        elif actual == expect.strip():
            status = base
            verdict = "OK"
        else:
            status = "FAIL"
            reason = "stdout mismatch vs @EXPECT"
            verdict = "MISMATCH"
    else:
        status = base

    return test_name, status, ai_used, reason, verdict, expect, actual, desc


def main():
    parser = argparse.ArgumentParser(description="Bee AI-aware test runner")
    parser.add_argument("target", nargs="?", default="level0",
                        help="level (e.g. level1) or a single test name (T0001)")
    parser.add_argument("--nv", action="store_true",
                        help="non-verbose: print only AI verdicts, not full IO")
    args = parser.parse_args()

    target = args.target
    verbose = not args.nv

    if os.path.exists(target) and target.endswith(".bee"):
        paths = [target]
        lvl_dir = os.path.dirname(target)
    elif os.path.isdir(os.path.join("test", target)):
        lvl_dir = os.path.join("test", target)
        paths = [os.path.join(lvl_dir, f)
                 for f in sorted(os.listdir(lvl_dir)) if f.endswith(".bee")]
    else:
        # Single named case - search all levels.
        for lvl in LEVELS:
            p = os.path.join("test", lvl, f"{target}.bee")
            if os.path.exists(p):
                paths = [p]
                lvl_dir = os.path.join("test", lvl)
                break
        else:
            print(f"Error: test case '{target}' not found.")
            sys.exit(2)

    passed, failed, skipped, ai_yes, ai_no = 0, 0, 0, 0, 0
    results = []

    for path in paths:
        test_name = os.path.splitext(os.path.basename(path))[0]
        name, status, ai_used, reason, verdict, expect, actual, desc = run_case(
            test_name, path, verbose)
        ai_col = "Yes" if ai_used else "No"

        if ai_used:
            ai_yes += 1
        else:
            ai_no += 1

        if status == "PASS":
            passed += 1
        elif status == "SKIP":
            skipped += 1
        else:
            failed += 1

        if ai_used:
            # Machine-readable verdict so an AI can confirm printed output.
            print(f"[{name}] AI=Yes expect=\"{esc(expect)}\" actual=\"{esc(actual)}\" "
                  f"-> {verdict}" + (f" ({reason})" if reason else ""))
        else:
            print(f"[{name}] AI=No status={status}"
                  + (f" ({reason})" if reason else ""))

        results.append({
            "name": test_name, "status": status, "ai": ai_col,
            "reason": reason, "verdict": verdict,
            "expect": expect, "actual": actual,
        })

        # Refresh the level README table (adds/updates the AI and status cols),
        # keeping the @DESC header text as the human + AI-readable description.
        status_arg = "FAIL" if status == "FAIL" else ("SKIP" if status == "SKIP" else "PASS")
        subprocess.run(["python", "scripts/update_test_readme.py",
                        test_name, status_arg, desc, ai_col], check=False)

    print("\n===============================")
    print(f"Level: {os.path.basename(lvl_dir)}  "
          f"Passed: {passed}  Failed: {failed}  Skipped: {skipped}")
    print(f"AI based (print): {ai_yes}   Assertion-only (no print): {ai_no}")

    status_dir = "test/status"
    os.makedirs(status_dir, exist_ok=True)
    ts = datetime.datetime.now().strftime("%Y%m%d_%H%M%S")
    with open(os.path.join(status_dir, f"ai_{ts}.json"), "w", encoding="utf-8") as sf:
        json.dump({"passed": passed, "failed": failed, "skipped": skipped,
                   "ai_yes": ai_yes, "ai_no": ai_no,
                   "results": results}, sf, indent=2)

    sys.exit(0 if failed == 0 else 1)


if __name__ == "__main__":
    main()
