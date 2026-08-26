import subprocess
import os

def test_execution(bee_path, test_path):
    print(f"Running execution test: {test_path}")
    result = subprocess.run([bee_path, "-e", test_path], capture_output=True, text=True)
    
    # We expect '30' and '20' to appear in the output.
    output = result.stdout.strip()
    print(f"Output received:\n{output}")
    
    if "30" in output and "20" in output:
        print("PASSED: Correct evaluation.")
    else:
        print("FAILED: Output did not match expected values.")

if __name__ == "__main__":
    bee_path = "./bin/bee.exe"
    test_path = "./test/level2/T0204.bee"
    test_execution(bee_path, test_path)
