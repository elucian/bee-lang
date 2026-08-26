#!/usr/bin/env python3
import subprocess
import sys

def test():
    print("Running Test Pipeline...")
    result = subprocess.run([sys.executable, "test/test_runner.py"])
    if result.returncode == 0:
        print("Tests passed.")
    else:
        print("Tests failed.")
        sys.exit(1)

if __name__ == "__main__":
    test()
