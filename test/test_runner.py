import os
import subprocess
import sys

def run_tests():
    print("Running Bee Compiler Tests...")
    test_dir = "."
    output_dir = "output"
    success = True
    
    # Simple runner: Find all .bee files in tests directory
    for root, dirs, files in os.walk(test_dir):
        for file in files:
            if file.endswith(".bee"):
                path = os.path.join(root, file)
                print(f"Testing {path}...")
                
                # Assume compiler exists at ./bee
                # Replace with actual compiler invocation
                result = subprocess.run(["./bee", path], capture_output=True, text=True)
                
                if result.returncode != 0:
                    print(f"FAILED: {path}")
                    print(result.stderr)
                    success = False
                else:
                    print(f"PASSED: {path}")

    if not success:
        sys.exit(1)
    print("All tests passed.")

if __name__ == "__main__":
    run_tests()
