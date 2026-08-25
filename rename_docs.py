import os

# Define the mapping of current names to new names
# Assuming 'doc' directory is in the current working directory or relative to it.
doc_dir = 'doc'

files_mapping = {
    'control.md': '06-control.md',
    'rules.md': '07-rules.md',
    'functions.md': '08-functions.md',
    'objects.md': '09-objects.md',
    'collections.md': '10-collections.md',
    'processing.md': '11-processing.md',
    'concurrency.md': '12-concurrency.md',
    'graphics.md': '13-graphics.md',
    'library.md': '14-library.md'
}

def rename_files():
    for current_name, new_name in files_mapping.items():
        src = os.path.join(doc_dir, current_name)
        dst = os.path.join(doc_dir, new_name)
        
        if os.path.exists(src):
            os.rename(src, dst)
            print(f"Renamed: {src} -> {dst}")
        else:
            print(f"File not found, skipping: {src}")

if __name__ == "__main__":
    rename_files()
