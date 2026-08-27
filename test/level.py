import os
import subprocess
import sys
import argparse
import json
import datetime

def run_tests():
    parser = argparse.ArgumentParser(description="Bee Language Level Test Runner")
    parser.add_argument("--level", type=int, help="Specify level to test (e.g. 1, 2, 3, 4, 5)")
    args = parser.parse_args()

    test_dir = "test"
    output_dir = "test/output"
    status_dir = "test/status"
    os.makedirs(output_dir, exist_ok=True)
    os.makedirs(status_dir, exist_ok=True)
        
    failed_count = 0
    passed_count = 0
    
    for root, dirs, files in os.walk(test_dir):
        if "output" in root or "bench" in root or "status" in root:
            continue
            
        if args.level:
            level_str = f"level{args.level}"
            if level_str not in root:
                continue

        for file in files:
            if file.endswith(".bee"):
                test_path = os.path.join(root, file)
                rel_path = os.path.relpath(test_path, test_dir)
                
                result = subprocess.run(["./bin/bee.exe", "-e", test_path], capture_output=True, text=True)
                
                if result.returncode != 0:
                    failed_count += 1
                    print(f"Testing {test_path}...\n  -> FAILED")
                    test_name_safe = rel_path.replace(os.sep, "_")
                    fail_output_path = os.path.join(output_dir, f"{test_name_safe}.fail")
                    with open(fail_output_path, "w") as f_out:
                        f_out.write(f"Test File: {test_path}\n")
                        f_out.write(f"--- STDOUT ---\n{result.stdout}\n")
                        f_out.write(f"--- STDERR (Errors) ---\n{result.stderr}\n")
                else:
                    passed_count += 1
                    print(f"Testing {test_path}...\n  -> PASSED")
    
    timestamp = datetime.datetime.now().strftime("%Y%m%d_%H%M%S")
    status_file = os.path.join(status_dir, f"status_{timestamp}.json")
    status_data = {
        "timestamp": timestamp,
        "level_filter": args.level if args.level else "all",
        "passed": passed_count,
        "failed": failed_count,
        "status": "SUCCESS" if failed_count == 0 else "FAILURE"
    }
    with open(status_file, "w") as f:
        json.dump(status_data, f, indent=2)
    print(f"Status report saved to {status_file}")

    print(f"\nTest run completed. Passed: {passed_count}, Failed: {failed_count}")
    if failed_count > 0:
        sys.exit(1)
    print("All tests passed.")

if __name__ == "__main__":
    run_tests()
