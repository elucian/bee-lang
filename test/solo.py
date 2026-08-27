# test/solo.py - Single test runner with single-attempt retry on failure
import os
import sys
import subprocess

def run_test(test_path, attempt_label=""):
    print(f"\n=== Running {test_path} {attempt_label} ===")
    res = subprocess.run(["./bin/bee.exe", "-e", "-d", test_path], capture_output=True, text=True)
    print("--- STDOUT ---")
    print(res.stdout)
    print("--- STDERR ---")
    print(res.stderr)
    return res.returncode, res.stdout, res.stderr

def run_solo():
    if len(sys.argv) < 2:
        print("Usage: python test/solo.py <test_case_name_or_path>")
        print("Example: python test/solo.py T0106")
        sys.exit(1)
        
    target = sys.argv[1]
    
    # Locate test file
    test_path = None
    if os.path.exists(target):
        test_path = target
    else:
        for lvl in ["level1", "level2", "level3", "level4", "level5", "critical"]:
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
        
    print("=== Rebuilding compiler ===")
    build_res = subprocess.run(["python", "build.py"])
    if build_res.returncode != 0:
        print("Build failed.")
        sys.exit(1)
        
    code, stdout, stderr = run_test(test_path, "(Attempt 1)")
    
    if code == 0:
        print(f"\n-> Test '{test_path}' PASSED successfully.")
        sys.exit(0)
    else:
        print(f"\n-> Attempt 1 failed. Rebuilding compiler and trying once more...")
        
        build_res = subprocess.run(["python", "build.py"])
        if build_res.returncode != 0:
            print("Build failed.")
            sys.exit(1)
            
        code2, stdout2, stderr2 = run_test(test_path, "(Attempt 2 - Final)")
        
        if code2 == 0:
            print(f"\n-> Test '{test_path}' PASSED on Attempt 2.")
            sys.exit(0)
        else:
            print(f"\n-> Test '{test_path}' FAILED twice. Stopping execution.")
            
            # Save fail report
            os.makedirs("test/output", exist_ok=True)
            fail_name = os.path.basename(test_path) + ".fail"
            fail_path = os.path.join("test/output", fail_name)
            
            code_lines = []
            try:
                with open(test_path, "r") as tf:
                    code_lines = tf.readlines()
            except Exception:
                pass
                
            with open(fail_path, "w") as f_out:
                f_out.write(f"Test File: {test_path}\n")
                f_out.write("--- STDOUT (Attempt 2) ---\n")
                f_out.write(stdout2 + "\n")
                f_out.write("--- STDERR (Attempt 2) ---\n")
                f_out.write(stderr2 + "\n")
                f_out.write("--- Source Code Outline ---\n")
                for idx, line in enumerate(code_lines, 1):
                    f_out.write(f"{idx:4d}\t{line}")
                    
            print(f"Failure report saved to {fail_path}")
            sys.exit(2)

if __name__ == "__main__":
    run_solo()
