import os
import re

def clean_doc_file(filepath):
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()

    # Regex to remove:
    # 1. ##### Lab Topics at the start
    # 2. * * * lines
    # 3. Copyright/Bottom garbage
    
    # Remove header metadata
    content = re.sub(r'^##### Lab Topics\s*\n\s*\*\s*\*\s*\*\s*\n\n*', '', content)
    
    # Remove trailing horizontal rules, copyright, and symbols
    content = re.sub(r'\n\n*\*\s*\*\s*\*\s*\n\n*©.*?\n\n*☰\s*$', '', content, flags=re.DOTALL)
    
    with open(filepath, 'w', encoding='utf-8') as f:
        f.write(content.strip() + '\n')

doc_dir = 'doc'
for filename in os.listdir(doc_dir):
    if filename.endswith('.md') and filename != 'index.md':
        print(f"Cleaning {filename}...")
        clean_doc_file(os.path.join(doc_dir, filename))
