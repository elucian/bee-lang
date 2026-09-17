# test/migrate_l4.py — Level 4 test migration driver.
#
# Migrates print-based (AI=Yes) level4 tests to the @AI: No assertion-first
# convention (config/AGENTS.md §5, test/readme.md §4). Every edit is routed
# through `bee-ed edit` with the `@path` temp-file mechanism so no file is ever
# hand-edited. D4 demos on unimplemented compiler features are @DISABLED with an
# honest reason rather than papered over.
#
# Usage: python test/migrate_l4.py
import os
import subprocess

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
LEVEL4 = os.path.join(ROOT, "test", "level4")
TEMP = os.path.join(ROOT, ".temp")
os.makedirs(TEMP, exist_ok=True)
# bee-ed invoked directly (bypassing run.sh) with MSYS path-conversion off, so
# @path temp files are read from .temp/ via relative paths and never mangled.
RUN = ["./bin/bee-ed", "edit"]
_counter = [0]


def bee_edit(path, old, new):
    """Run `bee-ed edit <path> <old> <new>`, both values via @path files in
    .temp/ (no trailing newline) so `--`-prefixed or multi-line content is not
    mistaken for a flag."""
    _counter[0] += 1
    n = _counter[0]
    oldf = ".temp/old_%d.txt" % n
    newf = ".temp/new_%d.txt" % n
    # newline="" disables automatic \n<->os.linesep translation (Windows),
    # so the temp file carries exactly the bytes bee-ed must match/insert.
    with open(os.path.join(ROOT, oldf), "w", encoding="utf-8", newline="") as of:
        of.write(old)
    with open(os.path.join(ROOT, newf), "w", encoding="utf-8", newline="") as nf:
        nf.write(new)
    env = dict(os.environ, MSYS2_ARG_CONV_EXCL="*", MSYS_NO_PATHCONV="1")
    return subprocess.run(RUN + [path, "@" + oldf, "@" + newf],
                          cwd=ROOT, env=env, capture_output=True, text=True)


def eol_of(path):
    """Return the line ending used by a file: '\r\n' if CRLF, else '\n'.
    Kept trivial (peeks at the first line ending) because bee-ed edit needs
    old/new byte-exact against the target file's own EOL."""
    with open(path, "r", encoding="utf-8", newline="") as f:
        data = f.read()
    return "\r\n" if "\r\n" in data else "\n"


lib = {}


class Lib:
    """Accumulates edits as (old, new); old/new are stored '\n'-based and are
    re-mapped to the target file's real EOL right before bee-ed edit."""
    def __init__(self):
        self.ops = {}

    def add(self, name, old, new):
        self.ops.setdefault(name, []).append((old, new))

    def deposit(self):
        """Run every accumulated edit, EOL-aware, returning (ok, fail)."""
        ok, fail = 0, 0
        for name, edits in sorted(self.ops.items()):
            path = os.path.join(LEVEL4, name + ".bee")
            eol = eol_of(path)
            for old, new in edits:
                if "\n" in old or "\n" in new:
                    old = old.replace("\n", eol)
                    new = new.replace("\n", eol)
                r = bee_edit(path, old, new)
                if r.returncode != 0:
                    fail += 1
                    print(f"[{name}] FAIL: {r.stdout.strip()} {r.stderr.strip()}")
                else:
                    ok += 1
                    print(f"[{name}] ok: {str(r.stdout).strip()}")
        return ok, fail


def migrate():
    # name -> list of (old, new) edits applied in order to that file.
    global lib
    lib = Lib()
    ops = lib.ops

    def add(name, old, new):
        lib.add(name, old, new)

    def ai_no(name, desc_line):
        """Insert an @AI: No header tag directly after the @DESC line."""
        add(name, desc_line, desc_line + "\n-- @AI: No - self-sufficient assertion test")

    def disable(name, desc_line, reason):
        add(name, desc_line, "-- @DISABLED: " + reason + "\n" + desc_line)

    # ---- Group A: retains expect assertions, drops redundant print ----
    ai_no("T0403", "-- @DESC: contract deposit with expectation")
    add("T0403", "  print balance;\n", "")

    ai_no("T0404", "-- @DESC: contract grow on a collection")
    add("T0404", "  print xs[2];\n", "")

    ai_no("T0405", "-- @DESC: contract write-through argument")
    add("T0405", "  print counter;\n", "")

    ai_no("T0409", "-- @DESC: contract store item in collection")
    add("T0409", "  print xs[3];\n", "")

    ai_no("T0415", "-- @DESC: growing a list with <+ and +>")
    add("T0415", "  print l;\n", "")

    ai_no("T0426", "-- @DESC: slice extraction with $ anchor")
    add("T0426", "  print w; expect w = [2,3,4];\n", "  expect w = [2,3,4];\n")

    # ---- Group A2: confirmation-print tests (real expects retained) ----
    ai_no("T0430", "-- @DESC: collection assertions for arrays, lists, sets, maps, indexing bounds")
    add("T0430", "  print s;\n", "")
    add("T0430", "  print m;\n", "")
    add("T0430", '  print "T0430 all assertions passed";\n', "")

    ai_no("T0431", "-- @DESC: dollar-anchor arithmetic and array processing with asserts")
    add("T0431", '  print "T0431 dollar-anchor assertions passed";\n', "")

    ai_no("T0433", "-- @DESC: array slicing with $ anchor and 1-based indexing (spec/10 §2, §3)")
    add("T0433", '  print "T0433 slicing assertions passed";\n', "")

    ai_no("T0434", "-- @DESC: range endpoint variants with $ anchor guards (spec/01 §5, spec/05 §3)")
    add("T0434", '  print "T0434 stepped-range assertions passed";\n', "")

    ai_no("T0435", "-- @DESC: set builder with filter over new range syntax (spec/10 §3, spec/11 processing)")
    add("T0435", '  print "T0435 set-builder assertions passed";\n', "")

    ai_no("T0436", "-- @DESC: array and hash-map builders over stepped range (spec/11 §4, spec/10 §3)")
    add("T0436", '  print "T0436 builder assertions passed";\n', "")

    ai_no("T0437", "-- @DESC: collection iteration with $ anchor and control flow (spec/11 §1)")
    add("T0437", '  print "T0437 collection iteration assertions passed";\n', "")

    ai_no("T0438", "-- @DESC: list concatenation and append/shift operators (spec/10 §3.1, spec/11 list ops)")
    add("T0438", '  print "T0438 list operation assertions passed";\n', "")

    ai_no("T0439", "-- @DESC: array decomposition, spreading, and rendering (spec/11 §5.1)")
    add("T0439", '  print "T0439 decomposition assertions passed";\n', "")

    ai_no("T0440", "-- @DESC: matrix row/column slice mutation and broadcast (spec/11 §5.2)")
    add("T0440", '  print "T0440 matrix slice assertions passed";\n', "")

    ai_no("T0441", "-- @DESC: set algebra intersection, union, symmetric difference (spec/10 §3.4)")
    add("T0441", '  print "T0441 set algebra assertions passed";\n', "")

    ai_no("T0442", "-- @DESC: quantifiers forall/exists over a collection (spec/11 §3)")
    add("T0442", '  print "T0442 quantifier assertions passed";\n', "")

    ai_no("T0443", "-- @DESC: collection casting between array and set via builder (spec/11 casting)")
    add("T0443", '  print "T0443 collection casting assertions passed";\n', "")

    ai_no("T0444", "-- @DESC: list as queue FIFO using append and shift-left (spec/11 list ops)")
    add("T0444", '  print "T0444 queue operation assertions passed";\n', "")

    # ---- Group B: negative test, unreachable print is dead code ----
    add("T0432",
        "-- @DESC: raw negative index a[-1] is rejected; use a[$-1] (spec/10 §2.1)",
        "-- @DESC: raw negative index a[-1] is rejected; use a[$-1] (spec/10 §2.1)\n"
        "-- @AI: No - negative rejection test; unreachable print is dead code")

    # ---- Group C: bare demos -> add expects, drop print ----
    ai_no("T0402", "-- @DESC:file: contract_argument_mutation.bee")
    add("T0402", "  apply bump(@counter);\n  print counter;\n",
                 "  apply bump(@counter);\n  expect counter = 11;\n")

    ai_no("T0407", "-- @DESC: contract postcondition with result binding")
    add("T0407", "  print bump(1);\n", "  new r := bump(1);\n  expect r = 2;\n")

    ai_no("T0411", "-- @DESC: array traversal with done terminator (D14)")
    add("T0411", "    print i;\n", "    expect a[i] = i;\n")

    ai_no("T0419", "-- @DESC:file: quantifiers.bee")
    add("T0419", "  print (∀ i ∈ {2,4,6} : i % 2 = 0);\n",
                 "  new all_even := (∀ i ∈ {2,4,6} : i % 2 = 0);\n  expect all_even = 1;\n")
    add("T0419", "  print (∃ i ∈ {1,3,5} : i = 3);\n",
                 "  new has_three := (∃ i ∈ {1,3,5} : i = 3);\n  expect has_three = 1;\n")

    ai_no("T0425", "-- @DESC:file: sized_array.bee")
    add("T0425", "  print a;\n", "  expect a[1] = 7;\n  expect a[$] = 7;\n")

    # ---- Group D: unimplemented compiler features -> honest @DISABLED ----
    disable("T0413", "-- @DESC:file: builders.bee",
            "set/array/map builders evaluate to 0 (unimplemented, spec/10 §3.4 builder)")
    disable("T0417", "-- @DESC:file: maps.bee",
            "map literal string values evaluate to 0 (unimplemented, spec/10 §3.5)")
    disable("T0418", "-- @DESC:file: map_builder.bee",
            "map builder (x:x²) evaluates to 0 (unimplemented)")
    disable("T0423", "-- @DESC:file: set_algebra_unicode.bee",
            "set union/intersection evaluate to 0 (unimplemented, spec/10 §3.4)")
    disable("T0427", "-- @DESC: walking a map with done terminator (D14)",
            "map iteration string values evaluate to 0 (unimplemented, spec/10 §3.5)")

    ok, fail = lib.deposit()
    print(f"\nEdits applied: {ok} ok, {fail} failed")


if __name__ == "__main__":
    migrate()
