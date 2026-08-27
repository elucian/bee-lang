#!/bin/sh
# run.sh - Master CLI helper script for Bee compiler workflow

COMMAND="$1"
TARGET="$2"

if [ "$COMMAND" = "build" ]; then
    python3 build.py
elif [ "$COMMAND" = "clean" ]; then
    echo "Cleaning test outputs and telemetry..."
    python3 clean.py
elif [ "$COMMAND" = "test" ]; then
    if [ -n "$TARGET" ]; then
        echo "Running test suite for $TARGET..."
        python3 test/test.py "$TARGET"
    else
        echo "Running complete test pipeline..."
        python3 test/test.py
    fi
elif [ "$COMMAND" = "solo" ]; then
    if [ -z "$TARGET" ]; then
        echo "Error: Target required for solo. Example: sh run.sh solo T0104"
        exit 1
    fi
    python3 test/solo.py "$TARGET"
else
    echo "Usage: sh run.sh [build | test [level1] | solo <target> | clean]"
    exit 1
fi
