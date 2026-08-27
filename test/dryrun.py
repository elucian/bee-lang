import os
import subprocess
import sys

def verify_all_flags():
    os.makedirs("test/output", exist_ok=True)
    
    test_cases = [
        {"args": ["-c", "test/level1/T0101.bee"], "name": "compile_c"},
        {"args": ["--compile", "test/level1/T0101.bee"], "name": "compile_long"},
        {"args": ["-e", "test/level1/T0101.bee"], "name": "execute_e"},
        {"args": ["--execute", "test/level1/T0101.bee"], "name": "execute_long"},
        {"args": ["-b", "test/level1/T0101.bee"], "name": "beautify_b"},
        {"args": ["--beautify", "test/level1/T0101.bee"], "name": "beautify_long"},
    ]
    
    failed = 0
    for tc in test_cases:
        print(f"Dryrun testing flag(s): {' '.join(tc['args'])}...")
        result = subprocess.run(["go", "run", "cmd/bee/main.go"] + tc['args'],
                               capture_output=True,
                               text=True)
        if result.returncode == 0:
            print(f"  -> PASSED ({tc['name']})")
        else:
            print(f"  -> FAILED ({tc['name']}): {result.stderr.strip()}")
            failed += 1
            
    if failed > 0:
        print(f"\nDryrun verification completed with {failed} failures.")
        sys.exit(1)
    else:
        print("\nAll compiler CLI flags verified successfully via dryrun!")

if __name__ == "__main__":
    verify_all_flags()
