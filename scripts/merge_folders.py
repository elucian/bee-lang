import os
import shutil
import sys

def merge_directories(src_root, dst_root):
    """
    Moves files from src_root to dst_root, maintaining subfolder structure,
    and merging files if they already exist.
    """
    if not os.path.exists(src_root):
        print(f"Source directory '{src_root}' does not exist.")
        return

    if not os.path.exists(dst_root):
        os.makedirs(dst_root)

    for root, dirs, files in os.walk(src_root):
        # Determine the relative path to create the corresponding structure in dst_root
        rel_path = os.path.relpath(root, src_root)
        target_dir = os.path.join(dst_root, rel_path)

        if not os.path.exists(target_dir):
            os.makedirs(target_dir)

        for file in files:
            src_file = os.path.join(root, file)
            dst_file = os.path.join(target_dir, file)

            # Move the file. If dst_file exists, it is overwritten, 
            # effectively 'merging' by replacing.
            shutil.move(src_file, dst_file)
            print(f"Moved: {src_file} -> {dst_file}")

if __name__ == "__main__":
    # Define paths relative to the project root
    src = "demo"
    dst = "test"
    
    merge_directories(src, dst)
    print("Task Completed")
