# test/watch.py
import subprocess
import sys

def run_watch(test_file):
    print(f"--- Running watch on {test_file} ---")
    # Build
    build = subprocess.run(["go", "build", "-o", "bin/bee.exe", "./cmd/bee/main.go"], capture_output=True, text=True)
    if build.returncode != 0:
        print(f"Build failed:\n{build.stderr}")
        sys.exit(1)
    else:
        # Execute in debug mode
        result = subprocess.run(["./bin/bee.exe", "-d", test_file], capture_output=True, text=True)
        print(f"Tokens output:\n{result.stdout}")
        print(result.stderr)
        sys.exit(result.returncode)

if __name__ == "__main__":
    test_file = sys.argv[1] if len(sys.argv) > 1 else "level2/T0202.bee"
    run_watch(test_file)
