import os
import re

def migrate_bee_files():
    # Define regex patterns for replacements
    # 1. Single '=' in comparisons needs to become '=='
    # NOTE: This is tricky to do safely without a full parser,
    # but we can target common 'expect' or 'if' patterns.
    
    # Let's start with clear '!=' -> 'is not' and '==' -> 'is' for identity (per instructions)
    # Actually user said:
    # "fix all test cases that use single = to use =="
    # "instead of != use "is not""
    # "instead of current == use "is" (for identity)"
    
    # We need a robust enough script to avoid breaking assignments.
    # Assignments use := or = in declarations.
    
    # Let's proceed carefully.
    
    bee_files = []
    for root, dirs, files in os.walk("test"):
        for file in files:
            if file.endswith(".bee"):
                bee_files.append(os.path.join(root, file))

    for path in bee_files:
        with open(path, 'r', encoding='utf-8') as f:
            content = f.read()

        # 1. Replace '!=' with 'is not'
        content = content.replace("!=", "is not")
        
        # 2. Replace '==' with 'is' (identity)
        # Note: Order matters. If we already replaced !=, we are okay.
        # However, some might be ==, some might be =.
        content = content.replace("==", "is")
        
        # 3. Handle single '=' in comparisons
        # This is dangerous. Let's look for 'expect a = b' -> 'expect a == b'
        # Pattern: (expect|if|when)\s+([^:=]+)=(?!=)([^:=]+)
        # That's also complex.
        # Let's try simple global replacement if not part of := or declarations.
        # Actually, if we've defined that declarations use '=' now,
        # we might have to be very careful.
        
        # Let's perform a manual check for now.
        
        with open(path, 'w', encoding='utf-8') as f:
            f.write(content)

migrate_bee_files()
