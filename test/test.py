import os
import re
import subprocess
import sys
import json
import datetime
import argparse

PRINT_RE = re.compile(r"\bprint\b")


def decode_expect(raw):
    """Decode `\n`/`\t` escapes in an @EXPECT footer tag."""
    return raw.replace("\\n", "\n").replace("\\t", "\t").strip() if raw else None


def uses_print(lines):
    """True iff the source BODY emits stdout via `print` (comments never count)."""
    for line in lines:
        if line.lstrip().startswith("--"):
            continue
        if PRINT_RE.search(line):
            return True
    return False

# Ensure Python's stdout/stderr can emit Unicode (e.g. ≠, ≤, ≥, etc.) on
# Windows consoles (cp1252 by default) and other non-UTF-8 terminals.
try:
    sys.stdout.reconfigure(encoding="utf-8")
    sys.stderr.reconfigure(encoding="utf-8")
except (AttributeError, ValueError):
    pass

def test():
    parser = argparse.ArgumentParser(description="Bee Test Orchestrator")
    parser.add_argument("level", nargs="?", help="Optional level to test (e.g. level1)")
    args = parser.parse_args()

    print("=== STEP: Running Test Cases ===")
    levels = [args.level] if args.level else ["level0", "level1", "level2", "level3", "level4", "level5", "level6", "level7", "level8"]
    level_results = {}
    total_passed = 0
    total_failed = 0
    
    for lvl in levels:
        lvl_dir = os.path.join("test", lvl)
        if not os.path.exists(lvl_dir):
            continue
            
        passed = 0
        failed = 0
        failed_cases = []
        
        for f in sorted(os.listdir(lvl_dir)):
            if f.endswith(".bee"):
                path = os.path.join(lvl_dir, f)
                disabled = False
                negative = False
                expect = None
                ai_tag = None
                desc = "TBD"
                with open(path, "r", encoding="utf-8") as tf:
                    lines = tf.readlines()
                # Footer-directive design (test/readme.md §4): scan the whole
                # file for lifecycle tags so footer @AI/@EXPECT/@NEGATIVE/@DISABLED are
                # always honored; @DESC is taken from its first (line 1) occurrence.
                for line in lines:
                    if "@DISABLED" in line:
                        disabled = True
                    if "@NEGATIVE" in line:
                        negative = True
                    if "-- @DESC:" in line and desc == "TBD":
                        desc = line.split("@DESC:")[1].strip()
                    if "-- @EXPECT:" in line:
                        expect = decode_expect(line.split("@EXPECT:")[1].strip())
                    m = re.search(r"@AI:\s*(Yes|No)", line)
                    if m:
                        ai_tag = m.group(1) == "Yes"
                if disabled:
                    continue
                test_name = os.path.splitext(f)[0]

                res = subprocess.run(["./bin/bee.exe", "-e", path], capture_output=True, text=True)

                print(f"--- STDOUT ({path}) ---")
                print(res.stdout)
                print(f"--- STDERR ({path}) ---")
                print(res.stderr)

                ai_used = uses_print(lines) if ai_tag is None else ai_tag
                if ai_used:
                    actual = res.stdout.strip()
                    if expect is None or expect == "":
                        status = "FAIL"
                        reason = "missing @EXPECT footer tag on print-based test"
                    elif actual != expect.strip():
                        status = "FAIL"
                        reason = "stdout mismatch vs @EXPECT"
                    else:
                        status = "PASS"
                        reason = ""
                elif negative:
                    status = "PASS" if res.returncode != 0 else "FAIL"
                else:
                    status = "PASS" if res.returncode == 0 else "FAIL"
                ai_col = "Yes" if ai_used else "No"
                subprocess.run(["python", "scripts/update_test_readme.py",
                                test_name, status, desc, ai_col])

                # Process report
                code_lines = lines
                
                output_dir = "test/output"
                os.makedirs(output_dir, exist_ok=True)
                report_path = os.path.join(output_dir, f"{test_name}.md")
                
                with open(report_path, "w", encoding="utf-8") as f_out:
                    f_out.write(f"# Test Execution Report: `{path}`\n\n")
                    f_out.write(f"--- STDOUT ---\n```\n{res.stdout}\n```\n\n")
                    f_out.write(f"--- STDERR ---\n```\n{res.stderr}\n```\n\n")
                    f_out.write("--- Source Code ---\n```bee\n")
                    f_out.writelines(code_lines)
                    f_out.write("```\n\n--- Conclusion ---\nStatus: " + status + "\n")

                if negative:
                    if res.returncode != 0:
                        passed += 1
                    else:
                        failed += 1
                        failed_cases.append(test_name)
                else:
                    if res.returncode == 0:
                        passed += 1
                    else:
                        failed += 1
                        failed_cases.append(test_name)

        level_results[lvl] = {"passed": passed, "failed": failed}
        total_passed += passed
        total_failed += failed
        
    status_dir = "test/status"
    os.makedirs(status_dir, exist_ok=True)
    timestamp = datetime.datetime.now().strftime("%Y%m%d_%H%M%S")
    with open(os.path.join(status_dir, f"status_{timestamp}.json"), "w") as sf:
        json.dump({"total_passed": total_passed, "total_failed": total_failed}, sf, indent=2)
        
    print(f"\nTest run finished. Total Passed: {total_passed}, Total Failed: {total_failed}")

if __name__ == "__main__":
    test()
