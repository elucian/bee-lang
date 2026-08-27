# build.py - Build Bee Compiler
import subprocess
import sys
import os

def build():
    print("Building Bee Compiler...")
    if not os.path.exists("bin"):
        os.makedirs("bin")
    result = subprocess.run(["go", "build", "-o", "bin/bee.exe", "./cmd/bee/main.go"])
    if result.returncode == 0:
        print("Build successful: bin/bee.exe")
    else:
        print("Build failed.")
        sys.exit(1)

if __name__ == "__main__":
    build()
