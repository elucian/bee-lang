import os
import re
import argparse

# Level README status tables carry four columns: CASE | DESCRIPTION | AI | STATUS.
# The AI column is Yes for print/output-producing (AI based) tests and No for
# purely assertion-driven tests. (See test/readme.md §4 and test/ai.py.)

# Human-level metadata used to bootstrap a missing/empty level README. The
# description and spec references mirror `test/readme.md` §1. Keys are the
# numeric level (0..8).
LEVEL_INFO = {
    0: ("Compiler Self-Bootstrap & Smoke", None),
    1: ("Lexical and Type Foundation", ["spec/01-lexical-structure.md",
                                        "spec/05-types.md"]),
    2: ("Statements and Control Flow", ["spec/02-statements.md",
                                        "spec/04-structure.md"]),
    3: ("Rules and Functions", ["spec/03-rules.md", "spec/07-functions.md"]),
    4: ("Collections and Pipelines", ["spec/10-collections.md",
                                      "spec/11-processing.md"]),
    5: ("Objects and Concurrency", ["spec/06-objects.md",
                                    "spec/12-concurrency.md"]),
    6: ("Advanced Features", None),
    7: ("System Integration", ["spec/14-library.md"]),
    8: ("Experimental", None),
}

PRINT_RE = re.compile(r"\bprint\b")


def level_number(test_name):
    """Map a test_name to its numeric level (0..8). `smoke` -> level0."""
    if test_name == "smoke":
        return 0
    return int(test_name[1:3])


def level_title(level_num):
    """Full first-line title for a level README, e.g. 'Test Level 4: ...'."""
    if level_num == 0:
        return "# Level 0: Compiler Self-Bootstrap & Smoke"
    blurb = LEVEL_INFO.get(level_num, ("Level X", None))[0]
    return f"# Test Level {level_num}: {blurb}"


def uses_print(path):
    """True iff the SOURCE BODY emits stdout via the `print` statement.

    Header comment lines (`-- @DESC:`, `-- @AI:`, ...) never count, matching
    test/ai.py::uses_print so the AI column stays consistent with the runner.
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


def scan_level(level_dir):
    """Scan a level dir and return per-test metadata.

    Returns a list of tuples:
      (test_name, description, ai("Yes"/"No"), status)
    `status` is derived honestly without executing anything:
      - SKIP if the header carries @DISABLED (feature not implemented yet)
      - UNRUN otherwise (not yet executed through the harness)
    `ai` reflects explicit @AI: Yes/No and otherwise body `print` usage.
    """
    rows = []
    if not os.path.isdir(level_dir):
        return rows
    for f in sorted(os.listdir(level_dir)):
        if not f.endswith(".bee"):
            continue
        path = os.path.join(level_dir, f)
        test_name = os.path.splitext(f)[0]
        desc, disabled, ai_tag = "see @DESC header", False, None
        try:
            with open(path, "r", encoding="utf-8") as tf:
                for line in tf:
                    if "@DISABLED" in line:
                        disabled = True
                    if "-- @DESC:" in line:
                        desc = line.split("@DESC:")[1].strip()
                    m = re.search(r"@AI:\s*(Yes|No)", line)
                    if m:
                        ai_tag = m.group(1) == "Yes"
        except Exception:
            pass
        if ai_tag is None:
            ai = "Yes" if uses_print(path) else "No"
        else:
            ai = "Yes" if ai_tag else "No"
        status = "SKIP" if disabled else "UNRUN"
        rows.append((test_name, desc, ai, status))
    return rows


def bootstrap_level_readme(level_num, readme_path):
    """Write a fresh, fully-populated level README from a scan of its .bee files.

    Used when the README is missing or effectively empty. Existing populated
    READMEs are left untouched (the per-test row update path handles them).
    """
    level_dir = f"test/level{level_num}"
    title = level_title(level_num)
    blurb = LEVEL_INFO.get(level_num, ("", None))[0]
    spec_refs = LEVEL_INFO.get(level_num, ("", None))[1]

    lines = [title + "\n", "\n"]
    if spec_refs:
        lines.append(f"This level covers {blurb.lower()} as defined in:\n")
        for ref in spec_refs:
            lines.append(f"- `{ref}`\n")
        lines.append("\n")
    else:
        lines.append(f"This level covers {blurb.lower()}.\n\n")
    lines.append("## Test Coverage\n")
    lines.append("| CASE     | DESCRIPTION                      | AI    | STATUS     |\n")
    lines.append("| -------- | -------------------------------- | ----- | ---------- |\n")

    for test_name, desc, ai, status in scan_level(level_dir):
        if len(desc) > 32:
            desc = desc[:29] + "..."
        lines.append(f"| {test_name:<8} | {desc:<32} | {ai:<6} | {status:<10} |\n")

    os.makedirs(level_dir, exist_ok=True)
    with open(readme_path, "w", encoding="utf-8") as f:
        f.writelines(lines)


def update_test_status(test_name, status, description, ai="No"):
    # Determine level from test_name (e.g., T0101 -> level1, T0001 -> level0)
    level_num = level_number(test_name)

    level_dir = f"test/level{level_num}"
    readme_path = os.path.join(level_dir, "README.md")

    if not os.path.exists(readme_path):
        print(f"Skipped {test_name}: no {readme_path}")
        return

    # Bootstrap a missing/empty README so `run.sh test`/`solo`/`ai` always leave
    # the level's README populated (header + full table), never empty.
    content = ""
    try:
        with open(readme_path, "r", encoding="utf-8") as f:
            content = f.read()
    except Exception:
        content = ""
    if not content.strip():
        bootstrap_level_readme(level_num, readme_path)

    with open(readme_path, "r", encoding="utf-8") as f:
        lines = f.readlines()

    # Column widths (mirrored by the table header/separator rows).
    w_case = 8
    w_desc = 32
    w_ai = 6
    w_stat = 10

    if len(description) > w_desc:
        description = description[:w_desc - 3] + "..."
    if not description or description == "TBD":
        description = "see @DESC header"
    ai = "Yes" if ai.lower() in ("yes", "1", "true") else "No"
    new_row = (f"| {test_name:<{w_case}} | {description:<{w_desc}} "
               f"| {ai:<{w_ai}} | {status:<{w_stat}} |\n")

    found = False
    for i, line in enumerate(lines):
        if line.strip().startswith('|'):
            parts = [p.strip() for p in line.split('|') if p.strip()]
            if parts and parts[0] == test_name:
                lines[i] = new_row
                found = True
                break

    if not found:
        # Insert after the separator row (`| --- |`).
        for i, line in enumerate(lines):
            if "---" in line:
                lines.insert(i + 1, new_row)
                break

    with open(readme_path, 'w', encoding='utf-8') as f:
        f.writelines(lines)
    print(f"Updated {test_name} in {readme_path} (status={status}, AI={ai})")


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("test_name")
    parser.add_argument("status")
    parser.add_argument("description", nargs="?", default="")
    parser.add_argument("ai", nargs="?", default="No")
    args = parser.parse_args()
    update_test_status(args.test_name, args.status, args.description, args.ai)
