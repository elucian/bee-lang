# clean.py - Cleanup Test Outputs
import os
import shutil

def clean():
    for d in ["test/output", "test/status", "test/critical"]:
        if os.path.exists(d):
            shutil.rmtree(d)
            os.makedirs(d)
            print(f"Cleaned {d} directory successfully.")
        else:
            os.makedirs(d, exist_ok=True)
            print(f"Created/verified {d} directory.")

if __name__ == "__main__":
    clean()
