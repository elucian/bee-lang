#!/bin/sh
# run.sh - Master CLI helper script for Bee compiler workflow

COMMAND="$1"
TARGET="$2"

if [ "$COMMAND" = "build" ]; then
    echo "Building Bee compiler..."
    python build.py
elif [ "$COMMAND" = "clean" ]; then
    echo "Cleaning test outputs and telemetry..."
    python clean.py
elif [ "$COMMAND" = "test" ]; then
    if [ -n "$TARGET" ]; then
        echo "Running test suite for $TARGET..."
        python test.py "$TARGET"
    else
        echo "Running complete test pipeline..."
        python test.py
    fi
else
    echo "Usage: sh run.sh [build|test [level1]|fix level1|fix T0104|clean]"
    exit 1
fi
