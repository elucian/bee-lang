import os
import re

# Configuration
DOC_DIR = 'doc'
TEST_DIR = 'test'

# Regex to match code blocks (indented or code-fenced)
# Bee examples in the doc are typically indented by 4 spaces
CODE_BLOCK_REGEX = re.compile(r'^(?:    |\t)(.*)', re.MULTILINE)

def extract_and_create_tests():
    # Ensure test directory exists
    if not os.path.exists(TEST_DIR):
        os.makedirs(TEST_DIR)

    # Process each doc file
    for filename in os.listdir(DOC_DIR):
        if filename.endswith('.md'):
            # Extract number-topic from filename
            base_name = filename.replace('.md', '')
            
            # Create corresponding test directory
            test_subfolder = os.path.join(TEST_DIR, base_name)
            if not os.path.exists(test_subfolder):
                os.makedirs(test_subfolder)
            
            doc_path = os.path.join(DOC_DIR, filename)
            
            with open(doc_path, 'r', encoding='utf-8') as f:
                content = f.read()
                
            # Find all code blocks
            # Note: This is a heuristic based on the Bee documentation format
            # where code is indented by 4 spaces.
            blocks = CODE_BLOCK_REGEX.findall(content)
            
            if blocks:
                # Group lines back into blocks if they are adjacent
                # For simplicity, we create one file per doc file containing extracted examples
                test_file_path = os.path.join(test_subfolder, 'extracted_examples.bee')
                with open(test_file_path, 'w', encoding='utf-8') as tf:
                    tf.write(f"-- Auto-extracted tests from {filename}\n\n")
                    # Simple heuristic: join lines that were part of indented blocks
                    # This logic may need refinement based on exact doc structure
                    tf.write("\n".join(blocks))
                    
                print(f"Created tests in {test_file_path}")

if __name__ == '__main__':
    extract_and_create_tests()
