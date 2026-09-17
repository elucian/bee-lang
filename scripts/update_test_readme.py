import os
import argparse

# Level README status tables carry four columns: CASE | DESCRIPTION | AI | STATUS.
# The AI column is Yes for print/output-producing (AI based) tests and No for
# purely assertion-driven tests. (See test/readme.md §4 and test/ai.py.)

def update_test_status(test_name, status, description, ai="No"):
    # Determine level from test_name (e.g., T0101 -> level1, T0001 -> level0)
    # Handle smoke.bee as well.
    if test_name == "smoke":
        level_num = 0
    else:
        level_num = int(test_name[1:3])

    level_dir = f"test/level{level_num}"
    readme_path = os.path.join(level_dir, "README.md")

    if not os.path.exists(readme_path):
        print(f"Skipped {test_name}: no {readme_path}")
        return

    with open(readme_path, 'r', encoding='utf-8') as f:
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
