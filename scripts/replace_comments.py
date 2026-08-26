"""
Purpose: Replace comments starting with '#' or '**' with '--' in demo subfolders.
Responsibility: Recursively process example files in demo directories, identifying
                lines that begin with '#' or '**' comment syntax and converting them
                to standard Bee language single-line comment syntax ('--').
Core Architectural Strategy: Uses Python's regex engine to match leading whitespace
                             followed by '#' or '**' prefixes, transforming them safely
                             while preserving relative indentation, line endings, and file structure.
"""

import os
import sys
import re
import argparse


def replace_comments_in_line(line: str) -> str:
    """
    Replaces comment markers starting with '#' or '**' at the beginning of a line with '--'.

    Inputs:
        line (str): A single line of text from a source file (may include trailing newline).

    Outputs:
        str: The transformed line with comment prefixes converted to '--',
             or the original line if no matching comment prefix was found.
    """
    # Inline spy comment: Extract trailing line ending (\r\n or \n) to preserve original formatting
    line_ending = ""
    if line.endswith("\r\n"):
        line_ending = "\r\n"
        content_line = line[:-2]
    elif line.endswith("\n"):
        line_ending = "\n"
        content_line = line[:-1]
    else:
        content_line = line

    # Inline spy comment: Match leading indentation followed by '#' or '**'
    pattern = r'^(\s*)(#+|\*\*+)\s*(.*)$'
    match = re.match(pattern, content_line)
    if match:
        indent = match.group(1)
        content = match.group(3)
        if content:
            return f"{indent}-- {content}{line_ending}"
        else:
            return f"{indent}--{line_ending}"

    return line


def process_file(file_path: str, dry_run: bool = False) -> int:
    """
    Processes a single file to replace comment prefixes starting with '#' or '**'.

    Inputs:
        file_path (str): Path to the target file.
        dry_run (bool): If True, previews changes without writing to file.

    Outputs:
        int: The number of comment lines modified in the file.
    """
    # Inline spy comment: Read content with UTF-8 encoding
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            lines = f.readlines()
    except Exception as e:
        sys.stderr.write(f"Error reading {file_path}: {e}\n")
        return 0

    modified_lines = []
    changes_count = 0

    for line in lines:
        new_line = replace_comments_in_line(line)
        if new_line != line:
            changes_count += 1
        modified_lines.append(new_line)

    if changes_count > 0 and not dry_run:
        with open(file_path, 'w', encoding='utf-8') as f:
            f.writelines(modified_lines)

    return changes_count


def process_directory(dir_path: str, dry_run: bool = False) -> tuple[int, int]:
    """
    Recursively processes all files within a target directory.

    Inputs:
        dir_path (str): The directory path to scan.
        dry_run (bool): If True, previews changes without modifying files.

    Outputs:
        tuple[int, int]: (total_files_modified, total_comments_replaced)
    """
    total_files_modified = 0
    total_comments_replaced = 0

    if not os.path.exists(dir_path):
        sys.stderr.write(f"Directory not found: {dir_path}\n")
        return 0, 0

    for root, _, files in os.walk(dir_path):
        for file in files:
            file_path = os.path.join(root, file)
            changes = process_file(file_path, dry_run=dry_run)
            if changes > 0:
                total_files_modified += 1
                total_comments_replaced += changes
                action = "Would modify" if dry_run else "Modified"
                print(f"{action}: {file_path} ({changes} comment(s) replaced)")

    return total_files_modified, total_comments_replaced


def main():
    """
    Main entry point for command-line execution.

    Inputs:
        None (parses sys.argv).

    Outputs:
        None (prints results to stdout/stderr).
    """
    parser = argparse.ArgumentParser(
        description="Replace comments starting with '#' or '**' with '--' in demo subfolders."
    )
    parser.add_argument(
        "path",
        nargs="?",
        default="demo",
        help="Target directory path to process (default: 'demo')"
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Preview changes without writing to disk"
    )

    args = parser.parse_args()

    print(f"Processing directory: {args.path} (dry-run={args.dry_run})")
    files_count, comments_count = process_directory(args.path, dry_run=args.dry_run)
    print(f"Finished. Total files modified: {files_count}, Total comments replaced: {comments_count}")


if __name__ == "__main__":
    main()
