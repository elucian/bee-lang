# test/test_runner.py
import os
import subprocess
import sys

def run_tests():
    print("Running Bee Compiler Tests...")
    test_dir = "test"
    output_dir = "test/output"
    
    if not os.path.exists(output_dir):
        os.makedirs(output_dir)
        
    success = True
    failed_count = 0
    passed_count = 0
    
    for root, dirs, files in os.walk(test_dir):
        if "output" in root:
            continue
            
        for file in files:
            if file.endswith(".bee"):
                test_path = os.path.join(root, file)
                rel_path = os.path.relpath(test_path, test_dir)
                
                # Execute the compiler in compile-only mode (-c)
                result = subprocess.run(["./bin/bee.exe", "-c", test_path], capture_output=True, text=True)
                
                if result.returncode != 0:
                    success = False
                    failed_count += 1
                    print(f"Testing {test_path}...\n  -> FAILED")
                    
                    # Spool out output only for failing tests
                    test_name_safe = rel_path.replace(os.sep, "_")
                    fail_output_path = os.path.join(output_dir, f"{test_name_safe}.fail")
                    
                    with open(fail_output_path, "w") as f_out:
                        f_out.write(f"Test File: {test_path}\n")
                        f_out.write(f"--- STDOUT ---\n{result.stdout}\n")
                        f_out.write(f"--- STDERR (Errors) ---\n{result.stderr}\n")
                    print(f"     Spooled failure report to {fail_output_path}")
                else:
                    passed_count += 1
                    print(f"Testing {test_path}...\n  -> PASSED")
    
    print(f"\nTest run completed. Passed: {passed_count}, Failed: {failed_count}")
    if not success:
        sys.exit(1)
    print("All tests passed.")

if __name__ == "__main__":
    run_tests()
