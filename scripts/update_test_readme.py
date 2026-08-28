import os
import argparse

def update_test_status(test_name, status, description):
    level_num = int(test_name[1:3])
    level_dir = f"test/level{level_num}"
    readme_path = os.path.join(level_dir, "README.md")
    
    if not os.path.exists(readme_path):
        return

    with open(readme_path, 'r', encoding='utf-8') as f:
        lines = f.readlines()

    # Column widths
    w_case = 8
    w_desc = 32
    w_stat = 10
    
    # Format new row
    if len(description) > w_desc:
        description = description[:w_desc-3] + "..."
    new_row = f"| {test_name:<{w_case}} | {description:<{w_desc}} | {status:<{w_stat}} |\n"

    found = False
    for i, line in enumerate(lines):
        if line.strip().startswith('|'):
            parts = [p.strip() for p in line.split('|') if p.strip()]
            if parts and parts[0] == test_name:
                lines[i] = new_row
                found = True
                break
    
    if not found:
        # If not found, insert after the separator row
        for i, line in enumerate(lines):
            if "---" in line:
                lines.insert(i + 1, new_row)
                break
                
    with open(readme_path, 'w', encoding='utf-8') as f:
        f.writelines(lines)
    print(f"Updated {test_name} in {readme_path}")

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("test_name")
    parser.add_argument("status")
    parser.add_argument("description")
    args = parser.parse_args()
    update_test_status(args.test_name, args.status, args.description)
