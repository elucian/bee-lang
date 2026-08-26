import os
import subprocess
import sys
import json
import datetime

def run_tests():
    print("Running Bee Compiler Tests...")
    test_dir = "test"
    output_dir = "test/output"
    
    if not os.path.exists(output_dir):
        os.makedirs(output_dir)
        
    success = True
    results = {"version": "0.1.0", "tests": []}
    
    for root, dirs, files in os.walk(test_dir):
        if "output" in root:
            continue
            
        for file in files:
            if file.endswith(".bee"):
                test_path = os.path.join(root, file)
                print(f"Testing {test_path}...")
                
                # Execute the compiler in compile-only mode (-c)
                result = subprocess.run(["./bin/bee.exe", "-c", test_path], capture_output=True, text=True)
                
                status = "PASSED" if result.returncode == 0 else "FAILED"
                if result.returncode != 0:
                    success = False
                
                results["tests"].append({
                    "path": test_path,
                    "status": status,
                    "stderr": result.stderr
                })
    
    report_file = os.path.join(output_dir, f"report_{datetime.datetime.now().strftime('%Y%m%d_%H%M%S')}.json")
    with open(report_file, 'w') as f:
        json.dump(results, f, indent=4)
        
    if not success:
        print("Tests failed. Check report.")
        sys.exit(1)
    print("All tests passed.")

if __name__ == "__main__":
    run_tests()
