import os
import re
import argparse

# Level README status tables carry four columns, in this order:
#   | CASE | AI | STATUS | DESCRIPTION |
# The AI column is Yes for print/output-producing (AI based) tests and No for
# purely assertion-driven tests. (See test/readme.md §4 and test/ai.py.)
#
# Layout is deliberately compact: only CASE/AI/STATUS are content-sized and the
# DESCRIPTION column is a FIXED width (DESC_WIDTH) and truncated to it with a
# trailing "...". If an existing README still uses the legacy column order
# (CASE | DESCRIPTION | AI | STATUS) the script DETECTS the mismatch on the
# next run and RECREATES the whole table in the canonical order.

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

# Canonical column order (uppercased) that every level table must match.
EXPECTED_COLUMNS = ["CASE", "AI", "STATUS", "DESCRIPTION"]

# Compact, fixed column widths. CASE/AI/STATUS fit their real content; the
# DESCRIPTION column is the only one that is padded to a fixed width and then
# truncated to it.
W_CASE = 5     # 'T0401'
W_AI = 3       # 'Yes' / 'No'
W_STAT = 6     # 'PASS' / 'FAIL' / 'SKIP' / 'UNRUN'
W_DESC = 25    # fixed width; longer descriptions are truncated with '...'

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


def file_metadata(path):
    """Return (description, ai, status) for a single .bee file.

    `status` is derived honestly without executing anything:
      - SKIP if the header carries @DISABLED (feature not implemented yet)
      - UNRUN otherwise (not yet executed through the harness)
    `ai` reflects explicit @AI: Yes/No and otherwise body `print` usage.
    """
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
    return desc, ai, status


def scan_level(level_dir):
    """Scan a level dir and return per-test (test_name, desc, ai, status)."""
    rows = []
    if not os.path.isdir(level_dir):
        return rows
    for f in sorted(os.listdir(level_dir)):
        if not f.endswith(".bee"):
            continue
        test_name = os.path.splitext(f)[0]
        desc, ai, status = file_metadata(os.path.join(level_dir, f))
        rows.append((test_name, desc, ai, status))
    return rows


def truncate_desc(desc):
    """Truncate a description to the fixed column width, trailing '...'."""
    if not desc or desc == "TBD":
        return "see @DESC header"
    if len(desc) > W_DESC:
        return desc[:W_DESC - 3] + "..."
    return desc


def table_header():
    return (f"| {'CASE':<{W_CASE}} | {'AI':<{W_AI}} | {'STATUS':<{W_STAT}} "
            f"| {'DESCRIPTION':<{W_DESC}} |\n")


def table_separator():
    return (f"| {'-'*W_CASE} | {'-'*W_AI} | {'-'*W_STAT} "
            f"| {'-'*W_DESC} |\n")


def row_line(case, ai, status, desc):
    return (f"| {case:<{W_CASE}} | {ai:<{W_AI}} | {status:<{W_STAT}} "
            f"| {truncate_desc(desc):<{W_DESC}} |\n")


def header_parts(line):
    return [p.strip() for p in line.split("|") if p.strip()]


def is_separator_row(parts):
    return all(p == "-" * len(p) for p in parts)


def table_matches_format(lines):
    """True iff the README's status table header uses the canonical column order."""
    for line in lines:
        if line.strip().startswith("|"):
            parts = [p.upper() for p in header_parts(line) if p]
            if parts and parts[0] == "CASE":
                return parts == EXPECTED_COLUMNS
    return False


def parse_rows(lines):
    """Parse an existing status table into {case: {COLUMN_NAME: value, ...}}.

    Column values are mapped through the table's own header, so rows in the
    legacy order (CASE | DESCRIPTION | AI | STATUS) are read correctly too.
    """
    rows = {}
    header = None
    for line in lines:
        if not line.strip().startswith("|"):
            continue
        parts = header_parts(line)
        if not parts:
            continue
        if header is None and parts[0] == "CASE":
            header = parts
            continue
        if header is None or is_separator_row(parts):
            continue
        case = parts[0]
        data = {}
        for name, value in zip(header, parts):
            data[name] = value
        rows[case] = data
    return rows


def build_table(level_dir, existing_lines):
    """Build the canonical table from existing rows, filling gaps from the .bee files."""
    rows = parse_rows(existing_lines)
    table = [table_header(), table_separator()]
    if not os.path.isdir(level_dir):
        return table
    for f in sorted(os.listdir(level_dir)):
        if not f.endswith(".bee"):
            continue
        case = os.path.splitext(f)[0]
        data = rows.get(case, {})
        path = os.path.join(level_dir, f)
        _, scan_ai, _ = file_metadata(path)
        desc = data.get("DESCRIPTION")
        if desc is None or desc == "TBD":
            desc, _, _ = file_metadata(path)
        ai = data.get("AI", scan_ai)
        status = data.get("STATUS", "UNRUN")
        table.append(row_line(case, ai, status, desc))
    return table


def splice_table(lines, new_table):
    """Replace the contiguous table block in `lines` with `new_table`."""
    start = end = None
    for i, line in enumerate(lines):
        if line.strip().startswith("|"):
            if start is None:
                start = i
            end = i
        elif start is not None:
            break
    if start is None:
        return lines + new_table
    return lines[:start] + new_table + lines[end + 1:]


def bootstrap_level_readme(level_num, readme_path):
    """Write a fresh, fully-populated level README from a scan of its .bee files."""
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
    lines += build_table(level_dir, [])

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

    with open(readme_path, "r", encoding="utf-8") as f:
        content = f.read()

    # Bootstrap a missing/empty README so `run.sh test`/`solo`/`ai` always leave
    # the level's README populated (header + full table), never empty.
    if not content.strip():
        bootstrap_level_readme(level_num, readme_path)
        with open(readme_path, "r", encoding="utf-8") as f:
            lines = f.readlines()
    else:
        lines = content.splitlines(keepends=True)

    # Self-healing table: if the existing table has the wrong column order,
    # recreate the ENTIRE table in the canonical order while preserving each
    # test's current status, then fall through to the per-row update below.
    if not table_matches_format(lines):
        print(f"{readme_path}: table columns out of date, recreating table")
        lines = splice_table(lines, build_table(level_dir, lines))

    ai = "Yes" if ai.lower() in ("yes", "1", "true") else "No"
    new_row = row_line(test_name, ai, status, description)

    found = False
    for i, line in enumerate(lines):
        if line.strip().startswith("|"):
            parts = [p.strip() for p in line.split("|") if p.strip()]
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

    with open(readme_path, "w", encoding="utf-8") as f:
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
