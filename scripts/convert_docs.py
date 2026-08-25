import os
import glob
from bs4 import BeautifulSoup
import html2text

def convert_html_to_md():
    input_dir = os.path.join("..", "doc", "bee")
    output_dir = os.path.join("..", "doc")
    
    # Define canonical order based on index.html
    topic_order = {
        "features": "01",
        "syntax": "02",
        "operators": "03",
        "structure": "04",
        "types": "05",
        "control": "06",
        "rules": "07",
        "functions": "08",
        "objects": "09",
        "collections": "10",
        "processing": "11",
        "concurrency": "12",
        "graphics": "13",
        "library": "14"
    }
    
    if not os.path.exists(output_dir):
        os.makedirs(output_dir)

    h = html2text.HTML2Text()
    h.ignore_links = False
    h.body_width = 0
    h.wrap_links = False
    h.pad_tables = True
    h.single_line_break = False
    h.mark_code = True

    print(f"Searching for HTML files in: {os.path.abspath(input_dir)}")
    files = glob.glob(os.path.join(input_dir, "*.html"))
    print(f"Found {len(files)} files.")

    for html_file in files:
        filename = os.path.basename(html_file)
        base_name = os.path.splitext(filename)[0]
        
        if filename in ["template.html", "index.html"]:
            continue
            
        prefix = topic_order.get(base_name)
        if not prefix:
            print(f"Skipping unknown topic: {filename}")
            continue
            
        output_filename = f"{prefix}-{base_name}.md"
            
        print(f"Converting: {filename} -> {output_filename}")
        with open(html_file, 'r', encoding='utf-8') as f:
            soup = BeautifulSoup(f, 'html.parser')
            
            # Format: Insert blank lines before <pre> and after </pre>
            for pre in soup.find_all('pre'):
                # Use a specific block structure
                code_content = pre.get_text()
                pre.replace_with(f"\n\n```\n{code_content}\n```\n\n")

            md_content = h.handle(str(soup))
            
            # Ensure consistent spacing
            md_content = md_content.replace('\n\n\n', '\n\n')
            
            output_path = os.path.join(output_dir, output_filename)
            with open(output_path, 'w', encoding='utf-8') as md_f:
                md_f.write(md_content)
        
        # Create index.md for the root doc folder
        index_md_path = os.path.join(output_dir, "index.md")
        with open(index_md_path, 'w', encoding='utf-8') as f:
            f.write("# Bee Programming Language Documentation\n\n")
            f.write("| # | Topic | Description |\n")
            f.write("|---|-------|-------------|\n")
        
            # Sort by prefix
            sorted_topics = sorted(topic_order.items(), key=lambda x: x[1])
            for base, prefix in sorted_topics:
                topic_name = base.replace('-', ' ').title()
                f.write(f"| {prefix} | [{topic_name}]({prefix}-{base}.md) | |\n")
    
        print(f"Created: {index_md_path}")

if __name__ == "__main__":
    convert_html_to_md()
