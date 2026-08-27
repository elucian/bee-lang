#!/bin/sh
# run.sh - Master CLI helper script for Bee compiler workflow

COMMAND="$1"
TARGET="$2"

if [ "$COMMAND" = "build" ]; then
    python build.py
elif [ "$COMMAND" = "clean" ]; then
    echo "Cleaning test outputs and telemetry..."
    python clean.py
elif [ "$COMMAND" = "test" ]; then
    if [ -n "$TARGET" ]; then
        echo "Running test suite for $TARGET..."
        python test/test.py "$TARGET"
    else
        echo "Running complete test pipeline..."
        python test/test.py
    fi
elif [ "$COMMAND" = "solo" ]; then
    python test/solo.py "$TARGET"
else
    echo "Usage: sh run.sh [build|test [level1]|fix level1|fix T0104|clean]"
    exit 1
fi
