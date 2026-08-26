# test.py - Run Test Pipeline
import subprocess
import sys

def test():
    print("Running Test Pipeline...")
    # Using 'python' which is the standard executable name on Windows
    result = subprocess.run(["python", "test/test_runner.py"])
    if result.returncode == 0:
        print("Tests passed.")
    else:
        print("Tests failed.")
        sys.exit(1)

if __name__ == "__main__":
    test()
