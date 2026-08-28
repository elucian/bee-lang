import os
import re
import sys

def reset_tests(level_target=""):
    test_dir = "test"
    levels = [level_target] if level_target else [f"level{i}" for i in range(1, 6)]
    
    pattern = re.compile(r"^//\s*@DISABLED:\s*.*$|^--\s*@DISABLED:\s*.*$")
    
    count = 0
    for lvl in levels:
        lvl_path = os.path.join(test_dir, lvl)
        if not os.path.isdir(lvl_path):
            continue
        for f in os.listdir(lvl_path):
            if f.endswith(".bee"):
                f_path = os.path.join(lvl_path, f)
                with open(f_path, "r", encoding="utf-8") as file:
                    lines = file.readlines()
                
                if lines and pattern.match(lines[0].strip()):
                    lines[0] = "// @ENABLED: Reset by user\n"
                    with open(f_path, "w", encoding="utf-8") as file:
                        file.writelines(lines)
                    print(f"Reset test: {f_path}")
                    count += 1
    print(f"Successfully reset {count} test files.")

if __name__ == "__main__":
    target = sys.argv[1] if len(sys.argv) > 1 else ""
    reset_tests(target)
