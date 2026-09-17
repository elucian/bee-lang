# clean.py - Cleanup Test Outputs
import os
import shutil

def clean_status():
    dirs_to_clean = ["test/status", "test/output", ".temp"]
    for d in dirs_to_clean:
        if os.path.exists(d):
            shutil.rmtree(d)
        os.makedirs(d)
        print(f"Cleaned {d} directory.")

if __name__ == "__main__":
    clean_status()
