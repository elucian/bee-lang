import os
import subprocess
import sys
import json
import datetime
import argparse

def test():
    parser = argparse.ArgumentParser(description="Bee Test Orchestrator")
    parser.add_argument("level", nargs="?", help="Optional level to test (e.g. level1)")
    args = parser.parse_args()

    if not args.level:
        print("=== STEP 1: Running CLI Dryrun Verification (`dryrun.py`) ===")
        dryrun_res = subprocess.run(["python", "test/dryrun.py"])
        if dryrun_res.returncode != 0:
            print("\n[STOPPED] Dryrun verification failed. Halting test pipeline.")
            sys.exit(1)
        print("-> Dryrun verification PASSED.\n")

        print("=== STEP 2: Running Benchmark Suite (`bench.py`) ===")
        bench_res = subprocess.run(["python", "test/bench.py"])
        if bench_res.returncode != 0:
            print("\n[STOPPED] Benchmark test suite failed. Halting test pipeline.")
            sys.exit(1)
        print("-> Benchmark test suite PASSED.\n")

    print("=== STEP 3: Running Test Cases ===")
    levels = [args.level] if args.level else ["level1", "level2", "level3", "level4", "level5"]
    level_results = {}
    total_passed = 0
    total_failed = 0
    
    for lvl in levels:
        lvl_dir = os.path.join("test", lvl)
        if not os.path.exists(lvl_dir):
            print(f"Error: Level directory '{lvl_dir}' does not exist.")
            continue
            
        passed = 0
        failed = 0
        failed_cases = []
        passed_cases = []
        
        for f in sorted(os.listdir(lvl_dir)):
            if f.endswith(".bee"):
                path = os.path.join(lvl_dir, f)
                test_name = os.path.splitext(f)[0]
                res = subprocess.run(["./bin/bee.exe", "-e", path], capture_output=True, text=True)
                if res.returncode == 0:
                    passed += 1
                    passed_cases.append(test_name)
                    print(f"Testing {path}...\n  -> PASSED")
                else:
                    failed += 1
                    failed_cases.append(test_name)
                    print(f"Testing {path}...\n  -> FAILED")
                    output_dir = "test/output"
                    os.makedirs(output_dir, exist_ok=True)
                    fail_output_path = os.path.join(output_dir, f"{lvl}_{test_name}.md")
                    
                    # Read file lines to extract failing expectation if possible
                    code_lines = []
                    try:
                        with open(path, "r") as tf:
                            code_lines = tf.readlines()
                    except Exception:
                        pass
                        
                    with open(fail_output_path, "w", encoding="utf-8") as f_out:
                        f_out.write(f"# Test Execution Report: `{path}`\n\n")
                        f_out.write(f"--- STDOUT ---\n```\n{res.stdout}\n```\n\n")
                        f_out.write(f"--- STDERR (Errors & Stack Trace) ---\n```\n{res.stderr}\n```\n\n")
                        f_out.write("--- Source Code Outline ---\n```bee\n")
                        for idx, line in enumerate(code_lines, 1):
                            clean_line = line.rstrip("\r\n")
                            if f"line {idx}" in res.stderr:
                                f_out.write(f"{idx:4d}\t{clean_line} -- **FAILED**\n")
                            else:
                                f_out.write(f"{idx:4d}\t{line}")
                        f_out.write("```\n\n--- Conclusion ---\nStatus: **TEST FAIL**\n")
                    
        level_results[lvl] = {
            "passed": passed,
            "failed": failed,
            "cases": failed_cases
        }
        total_passed += passed
        total_failed += failed
        
    status_dir = "test/status"
    os.makedirs(status_dir, exist_ok=True)
    timestamp = datetime.datetime.now().strftime("%Y%m%d_%H%M%S")
    status_file = os.path.join(status_dir, f"status_{timestamp}.json")
    
    report = {
        "timestamp": timestamp,
        "levels": level_results,
        "total_passed": total_passed,
        "total_failed": total_failed,
        "status": "SUCCESS" if total_failed == 0 else "FAILURE"
    }
    
    with open(status_file, "w") as sf:
        json.dump(report, sf, indent=2)
        
    print(f"\nTest run finished. Total Passed: {total_passed}, Total Failed: {total_failed}")
    print(f"Status report saved to {status_file}")
    
    if total_failed > 0:
        sys.exit(1)

if __name__ == "__main__":
    test()
