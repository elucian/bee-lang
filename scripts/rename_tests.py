import os
import re

def process_test_files(base_path):
    # Regex to match T0XNN format and extract level and number
    pattern = re.compile(r"T0(\d)(\d{2})\.bee")
    
    # Track the highest number for each level
    max_numbers = {}

    # First pass: find the highest existing number per level
    for root, _, files in os.walk(base_path):
        for file in files:
            match = pattern.match(file)
            if match:
                level = int(match.group(1))
                number = int(match.group(2))
                max_numbers[level] = max(max_numbers.get(level, 0), number)

    # Second pass: process and rename files
    for root, _, files in os.walk(base_path):
        # Extract level from folder name (assuming 'levelX' format)
        dir_match = re.search(r"level(\d)", os.path.basename(root))
        if not dir_match:
            continue
            
        level = int(dir_match.group(1))
        
        for file in files:
            # Skip files that are already in the correct format
            if pattern.match(file):
                continue
            
            if file.endswith(".bee"):
                file_path = os.path.join(root, file)
                
                # Increment number for this level
                max_numbers[level] = max_numbers.get(level, 0) + 1
                new_name = f"T0{level}{max_numbers[level]:02d}.bee"
                new_path = os.path.join(root, new_name)
                
                # Prepend comment
                with open(file_path, 'r', encoding='utf-8') as f:
                    content = f.read()
                
                new_content = f"-- file: {file}\n{content}"
                
                with open(file_path, 'w', encoding='utf-8') as f:
                    f.write(new_content)
                
                # Rename
                os.rename(file_path, new_path)
                print(f"Renamed: {file} -> {new_name}")

if __name__ == "__main__":
    process_test_files("test")
    print("Task Completed")
