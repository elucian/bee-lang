#!/bin/bash
# build.sh - Build Bee Compiler
echo "Building Bee Compiler..."
go build -o bee ./cmd/bee/main.go
if [ $? -eq 0 ]; then
    echo "Build successful: ./bee"
else
    echo "Build failed."
    exit 1
fi
