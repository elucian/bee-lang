import os
import subprocess
import sys
import json
import datetime

def test_flags():
    os.makedirs("test/output", exist_ok=True)
    
    test_cases = [
        {"args": ["-h"], "name": "help_flag"},
        {"args": ["-v"], "name": "version_flag"},
        {"args": ["-d", "test/level1/T0101.bee"], "name": "debug_lexer"},
        {"args": ["-c", "test/level1/T0101.bee"], "name": "compile_only"}
    ]
    
    for tc in test_cases:
        output_file = f"test/output/{tc['name']}.out"
        print(f"Testing flag: {' '.join(tc['args'])}...")
        
        result = subprocess.run(["./bin/bee.exe"] + tc['args'], 
                               capture_output=True, 
                               text=True)
        
        with open(output_file, "w") as f:
            f.write(result.stdout)
            if result.stderr:
                f.write(result.stderr)
        
        if os.path.getsize(output_file) > 0:
            print(f"PASSED: {tc['name']} produced output.")
        else:
            print(f"FAILED: {tc['name']} produced no output.")

if __name__ == "__main__":
    test_flags()
