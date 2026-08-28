import os
import re

def update_test_descriptions(base_dir):
    # Pattern to match: -- T0XXX: Description OR -- Description
    pattern_id = re.compile(r"^-- (T\d{4}): (.+)$")
    pattern_comment = re.compile(r"^-- (.+)$")
    
    for root, _, files in os.walk(base_dir):
        for file in files:
            if file.endswith(".bee"):
                path = os.path.join(root, file)
                
                with open(path, 'r', encoding='utf-8') as f:
                    lines = f.readlines()
                
                if not lines:
                    continue
                
                first_line = lines[0].strip()
                if "@DESC:" in first_line:
                    continue
                
                match_id = pattern_id.match(first_line)
                match_comment = pattern_comment.match(first_line)
                
                if match_id:
                    test_id = match_id.group(1)
                    desc = match_id.group(2)
                    lines[0] = f"-- @DESC:{test_id} {desc}\n"
                    with open(path, 'w', encoding='utf-8') as f:
                        f.writelines(lines)
                    print(f"Updated standard description in {file}")
                elif match_comment:
                    comment = match_comment.group(1)
                    lines[0] = f"-- @DESC:{comment}\n"
                    with open(path, 'w', encoding='utf-8') as f:
                        f.writelines(lines)
                    print(f"Added @DESC to comment in {file}")

if __name__ == "__main__":
    update_test_descriptions("test")
