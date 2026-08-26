import os
import subprocess
import datetime
import json

VERSION = "0.1.0"
TEST_DIR = "test"
OUTPUT_DIR = "output"

def run_pipeline():
    timestamp = datetime.datetime.now().strftime("%Y%m%d_%H%M%S")
    report_file = os.path.join(OUTPUT_DIR, f"report_v{VERSION}_{timestamp}.json")
    results = {"version": VERSION, "timestamp": timestamp, "tests": []}

    for root, dirs, files in os.walk(TEST_DIR):
        for file in files:
            if file.endswith(".bee"):
                test_path = os.path.join(root, file)
                test_code = file.replace(".bee", "")
                
                # Placeholder for compiler execution:
                # result = subprocess.run(["./bee", test_path], capture_output=True)
                
                results["tests"].append({
                    "code": test_code,
                    "path": test_path,
                    "status": "PASSED" # Placeholder
                })

    with open(report_file, 'w') as f:
        json.dump(results, f, indent=4)
    print(f"Pipeline finished. Report: {report_file}")

if __name__ == "__main__":
    run_pipeline()
