import os
import re
import sys
import argparse

def update_test_status(test_name, status, description):
    # Determine level from test_name (e.g., T0101 -> level1)
    level_num = int(test_name[1:3])
    level_dir = f"test/level{level_num}"
    readme_path = os.path.join(level_dir, "README.md")
    
    if not os.path.exists(readme_path):
        print(f"Error: README not found for level {level_num}")
        return

    # Truncate description
    truncated_desc = (description[:27] + '..') if len(description) > 29 else description

    with open(readme_path, 'r') as f:
        lines = f.readlines()

    new_lines = []
    found = False
    for line in lines:
        if test_name in line:
            # Pattern: | T0XXX.bee | OLD_STATUS | DESC | LINK |
            parts = line.split('|')
            if len(parts) >= 4:
                parts[2] = f" {truncated_desc} "
                parts[3] = f" {status} "
                line = '|'.join(parts) + '\n'
            found = True
        new_lines.append(line)

    if found:
        with open(readme_path, 'w') as f:
            f.writelines(new_lines)
        print(f"Updated {test_name} in {readme_path} to {status}")
    else:
        print(f"Test {test_name} not found in {readme_path}")

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("test_name")
    parser.add_argument("status")
    parser.add_argument("description")
    args = parser.parse_args()
    
    update_test_status(args.test_name, args.status, args.description)
