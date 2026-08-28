# clean.py - Cleanup Test Outputs
import os
import shutil

def clean_status():
    status_dir = "test/status"
    if os.path.exists(status_dir):
        shutil.rmtree(status_dir)
    os.makedirs(status_dir)
    print(f"Cleaned {status_dir} directory.")

if __name__ == "__main__":
    clean_status()
