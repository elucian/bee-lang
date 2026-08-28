# test/solo.py - Single test runner without build
import os
import sys
import subprocess

def run_solo():
    if len(sys.argv) < 2:
        print("Usage: python test/solo.py <test_case_name_or_path>")
        sys.exit(1)
        
    target = sys.argv[1]
    test_path = None
    if os.path.exists(target):
        test_path = target
    else:
        for lvl in ["level0", "level1", "level2", "level3", "level4", "level5", "level6", "level7", "level8"]:
            p = os.path.join("test", lvl, f"{target}.bee")
            if os.path.exists(p):
                test_path = p
                break
            p2 = os.path.join("test", lvl, target)
            if os.path.exists(p2):
                test_path = p2
                break
                
    if not test_path or not os.path.exists(test_path):
        print(f"Error: Test case '{target}' not found.")
        sys.exit(1)
        
    print(f"=== Running test: {test_path} ===")
    
    # Run subprocess with explicit utf-8 encoding for Windows/Unix compatibility
    res = subprocess.run(
        ["./bin/bee.exe", "-e", "-d", test_path],
        capture_output=True,
        text=True,
        encoding="utf-8",
        errors="replace"
    )
    
    print("--- STDOUT ---")
    print(res.stdout)
    print("--- STDERR ---")
    print(res.stderr)
    
    code_lines = []
    try:
        with open(test_path, "r", encoding="utf-8") as tf:
            code_lines = tf.readlines()
    except Exception:
        pass
        
    os.makedirs("test/output", exist_ok=True)
    report_name = f"{os.path.splitext(os.path.basename(test_path))[0]}.md"
    report_path = os.path.join("test/output", report_name)
    
    status = "PASS" if res.returncode == 0 else "FAIL"
    
    # Extract description
    desc = "TBD"
    try:
        with open(test_path, "r", encoding="utf-8") as tf:
            for line in tf:
                if "-- @DESC:" in line:
                    desc = line.split("@DESC:")[1].strip()
                    break
    except:
        pass
    
    # Update README
    test_name = os.path.splitext(os.path.basename(test_path))[0]
    subprocess.run(["python", "scripts/update_test_readme.py", test_name, status, desc])

    with open(report_path, "w", encoding="utf-8") as f_out:
        f_out.write(f"# Test Execution Report: `{test_path}`\n\n")
        f_out.write("## Standard Output (`STDOUT`)\n```\n" + res.stdout + "\n```\n\n")
        f_out.write("## Standard Error & Trace (`STDERR`)\n```\n" + res.stderr + "\n```\n\n")
        f_out.write("## Source Code Outline\n```bee\n")
        for idx, line in enumerate(code_lines, 1):
            clean_line = line.rstrip("\r\n")
            if f"panic: expect failed at line {idx}" in res.stderr:
                f_out.write(f"{idx:4d}\t{clean_line} -- FAILED\n")
            else:
                f_out.write(f"{idx:4d}\t{line}")
        f_out.write("```\n\n## Conclusion\nStatus: " + status + "\n")
        
    print(f"Report saved to {report_path}")
    sys.exit(res.returncode)

if __name__ == "__main__":
    run_solo()
