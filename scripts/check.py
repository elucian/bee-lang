import os
import subprocess
import sys
from datetime import datetime

def check_syntax(base_path):
    """
    Scans the test directory for all .bee files and runs the compiler 
    in a mode that only checks syntax without execution.
    Creates output and status reports.
    """
    compiler_cmd = "bee" 
    check_flag = "-c" 
    
    success = True
    
    # Ensure status directory exists
    status_dir = os.path.join("test", "status")
    os.makedirs(status_dir, exist_ok=True)
    
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    report_path = os.path.join(status_dir, f"report_{timestamp}.txt")
    status_path = os.path.join(status_dir, f"status_{timestamp}.txt")
    
    with open(report_path, "w") as report_file, open(status_path, "w") as status_file:
        report_file.write(f"Syntax Check Report - {datetime.now()}\n\n")
        
        for root, _, files in os.walk(base_path):
            # Skip status directory
            if "status" in root:
                continue
            for file in files:
                if file.endswith(".bee"):
                    file_path = os.path.join(root, file)
                    
                    try:
                        result = subprocess.run(
                            [compiler_cmd, check_flag, file_path],
                            capture_output=True,
                            text=True
                        )
                        
                        if result.returncode != 0:
                            msg = f"Syntax Error in {file_path}:\n{result.stderr}\n"
                            print(msg)
                            report_file.write(msg)
                            status_file.write(f"FAIL: {file_path} -critical\n")
                            success = False
                        else:
                            msg = f"Passed: {file_path}\n"
                            print(msg)
                            report_file.write(msg)
                            status_file.write(f"PASS: {file_path}\n")
                            
                    except FileNotFoundError:
                        print(f"Compiler '{compiler_cmd}' not found.")
                        return False
    
    print(f"Reports generated:\n- Output: {report_path}\n- Status: {status_path}")
    return success

if __name__ == "__main__":
    if check_syntax(os.path.join("test")):
        print("All tests passed syntax check.")
        sys.exit(0)
    else:
        print("Some tests failed syntax check.")
        sys.exit(1)
