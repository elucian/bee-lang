import os
import re

def populate_readme(readme_path):
    if not os.path.exists(readme_path):
        return

    level_dir = os.path.dirname(readme_path)
    
    # Read existing README
    with open(readme_path, 'r', encoding='utf-8') as f:
        lines = f.readlines()

    # Collect existing test status/desc
    known_tests = {}
    for line in lines:
        if line.strip().startswith('|'):
            parts = [p.strip() for p in line.split('|') if p.strip()]
            if len(parts) >= 3 and not "CASE" in parts[0]:
                known_tests[parts[0]] = (parts[1], parts[2])

    # Find all .bee files
    bee_files = sorted([f for f in os.listdir(level_dir) if f.endswith(".bee")])
    
    # Rebuild table
    new_lines = []
    w_case, w_desc, w_stat = 8, 32, 10
    header = f"| {'CASE':<{w_case}} | {'DESCRIPTION':<{w_desc}} | {'STATUS':<{w_stat}} |\n"
    separator = f"| {'-'*w_case} | {'-'*w_desc} | {'-'*w_stat} |\n"
    
    new_lines.append(header)
    new_lines.append(separator)
    
    for f in bee_files:
        test_name = os.path.splitext(f)[0]
        desc, status = known_tests.get(test_name, ("TBD", "UNKNOWN"))
        
        # If still TBD, try to extract description from file
        if desc == "TBD":
            path = os.path.join(level_dir, f)
            with open(path, 'r', encoding='utf-8') as tf:
                for line in tf:
                    if "-- @DESC:" in line:
                        desc = line.split("@DESC:")[1].strip()
                        break
        
        if len(desc) > w_desc:
            desc = desc[:w_desc-3] + "..."
            
        new_lines.append(f"| {test_name:<{w_case}} | {desc:<{w_desc}} | {status:<{w_stat}} |\n")

    # Reconstruct file
    final_lines = []
    in_table = False
    for line in lines:
        if line.strip().startswith('|'):
            if not in_table:
                final_lines.extend(new_lines)
                in_table = True
        else:
            if not in_table:
                final_lines.append(line)
            # if we were in table and hit non-table, we already added table
            in_table = False

    with open(readme_path, 'w', encoding='utf-8') as f:
        f.writelines(final_lines)
    print(f"Populated {readme_path}")

for i in range(1, 9):
    populate_readme(f"test/level{i}/README.md")
