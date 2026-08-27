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
        
        for f in sorted(os.listdir(lvl_dir)):
            if f.endswith(".bee"):
                path = os.path.join(lvl_dir, f)
                res = subprocess.run(["./bin/bee.exe", "-e", path], capture_output=True, text=True)
                if res.returncode == 0:
                    passed += 1
                    print(f"Testing {path}...\n  -> PASSED")
                else:
                    failed += 1
                    print(f"Testing {path}...\n  -> FAILED")
                    
        level_results[lvl] = {
            "passed": passed,
            "failed": failed
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
    print(f"\nStatus report saved to {status_file}")

    print(f"\nTest run finished. Total Passed: {total_passed}, Total Failed: {total_failed}")
    if total_failed > 0:
        print("Some tests failed.")
        sys.exit(1)
    else:
        print("All executed tests passed successfully!")
    
    print("\nWaiting for user intervention. Please decide the next move.")

if __name__ == "__main__":
    test()
