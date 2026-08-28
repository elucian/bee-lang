import os
import re

def reformat_readme(readme_path):
    if not os.path.exists(readme_path):
        return

    with open(readme_path, 'r', encoding='utf-8') as f:
        lines = f.readlines()

    # Column widths
    w_case = 8
    w_desc = 32
    w_stat = 10
    
    new_lines = []
    in_table = False
    
    # Header row
    header = f"| {'CASE':<{w_case}} | {'DESCRIPTION':<{w_desc}} | {'STATUS':<{w_stat}} |\n"
    separator = f"| {'-'*w_case} | {'-'*w_desc} | {'-'*w_stat} |\n"
    
    for line in lines:
        if line.strip().startswith('|'):
            if not in_table:
                new_lines.append(header)
                new_lines.append(separator)
                in_table = True
            
            if "CASE" in line or "---" in line or "Test File" in line:
                continue
                
            parts = [p.strip() for p in line.split('|') if p.strip()]
            if len(parts) >= 3:
                case = parts[0].replace('.bee', '')
                # Re-map correctly: existing tables had | CASE | STATUS | DESC
                # Our new structure is | CASE | DESC | STATUS
                desc = parts[2]
                status = parts[1]
                
                if len(desc) > w_desc:
                    desc = desc[:w_desc-3] + "..."
                
                new_lines.append(f"| {case:<{w_case}} | {desc:<{w_desc}} | {status:<{w_stat}} |\n")
        elif not line.strip().startswith("*(Note:"):
            in_table = False
            new_lines.append(line)

    with open(readme_path, 'w', encoding='utf-8') as f:
        f.writelines(new_lines)
    print(f"Reformatted {readme_path}")

# Run for all levels
for i in range(1, 9):
    reformat_readme(f"test/level{i}/README.md")
