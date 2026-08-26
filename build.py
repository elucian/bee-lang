# build.py - Build Bee Compiler
import subprocess
import sys

def build():
    print("Building Bee Compiler...")
    result = subprocess.run(["go", "build", "-o", "bee.exe", "./cmd/bee/main.go"])
    if result.returncode == 0:
        print("Build successful: bee.exe")
    else:
        print("Build failed.")
        sys.exit(1)

if __name__ == "__main__":
    build()
