# clean.py - Cleanup Test Outputs
import os
import shutil

def clean():
    output_dir = "test/output"
    if os.path.exists(output_dir):
        shutil.rmtree(output_dir)
        os.makedirs(output_dir)
        print("Cleaned test/output directory successfully.")
    else:
        print("test/output directory does not exist.")

if __name__ == "__main__":
    clean()
