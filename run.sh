#!/bin/sh
# run.sh - Master CLI helper script for Bee compiler workflow

COMMAND="$1"
TARGET="$2"

if [ "$COMMAND" = "build" ]; then
    python clean.py
    python build.py
elif [ "$COMMAND" = "clean" ] || [ "$COMMAND" = "clan" ]; then
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
    if [ -z "$TARGET" ]; then
        echo "Error: Target required for solo. Example: sh run.sh solo T0104"
        exit 1
    fi
    python test/solo.py "$TARGET"
elif [ "$COMMAND" = "reset" ]; then
    if [ -n "$TARGET" ]; then
        echo "Resetting disabled tests for $TARGET..."
        python test/reset.py "$TARGET"
    else
        echo "Resetting all disabled tests..."
        python test/reset.py
    fi
else
    echo "Usage: sh run.sh [build | test [level1] | solo <target> | reset [level1] | clean]"
    exit 1
fi
