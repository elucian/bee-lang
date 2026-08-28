import os
import subprocess
import sys

def smoke_test():
    print("=== STEP 1: Building Health Checker ===")
    # Ensure bin exists
    os.makedirs("bin", exist_ok=True)

    # We are running health.go as a self-contained health check
    health_path = "test/level0/health.go"
    print("\n=== STEP 1: Running Self Check ===")
    print(f"-> Running health check: {health_path}")
    res = subprocess.run(["go", "run", health_path], capture_output=True, text=True)

    if res.returncode != 0:
        print(f"Health check FAILED:\n{res.stderr}")
        sys.exit(1)
    print("-> Health check PASSED.")

    print("\n=== STEP 2: Running Level0 Tests ===")
    test_res = subprocess.run(["python", "test/test.py", "level0"])
    if test_res.returncode != 0:
        print("Level0 test suite FAILED.")
        sys.exit(1)

    print("\nSyntax OK.")
    sys.exit(0)


if __name__ == "__main__":
    smoke_test()
