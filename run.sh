#!/bin/sh
# run.sh - Master CLI helper script for Bee compiler workflow

COMMAND="$1"
TARGET="$2"

if [ "$COMMAND" = "build" ]; then
    echo "Building Bee compiler..."
    python build.py
elif [ "$COMMAND" = "test" ]; then
    if [ -n "$TARGET" ]; then
        echo "Running test suite for $TARGET..."
        python test.py "$TARGET"
    else
        echo "Running complete test pipeline..."
        python test.py
    fi
elif [ "$COMMAND" = "fix" ]; then
    if [ -z "$TARGET" ]; then
        echo "Error: Please specify a target level or test case (e.g., sh run.sh fix level1 or sh run.sh fix T0104)"
        exit 1
    fi

    case "$TARGET" in
        T[0-9][0-9][0-9][0-9])
            echo "Attempting automated iterative fix for test case $TARGET (max 5 iterations)..."
            TEST_FILE=""
            for lvl in level1 level2 level3 level4 level5 critical; do
                if [ -f "test/$lvl/$TARGET.bee" ]; then
                    TEST_FILE="test/$lvl/$TARGET.bee"
                    break
                fi
            done
            
            if [ -z "$TEST_FILE" ]; then
                echo "Error: Test case $TARGET.bee not found in any level directory."
                exit 1
            fi

            i=1
            while [ $i -le 5 ]; do
                echo "--- Iteration $i / 5 for test case $TARGET ---"
                python -c "import subprocess, sys; res = subprocess.run(['./bin/bee.exe', '-e', '$TEST_FILE'], capture_output=True, text=True); print(res.stdout); print(res.stderr); sys.exit(res.returncode)"
                if [ $? -eq 0 ]; then
                    echo "Success! Test case $TARGET passed on iteration $i."
                    exit 0
                fi
                echo "Test failed. Running compiler build and re-evaluating..."
                python build.py
                i=$((i + 1))
            done
            echo "Reached maximum 5 iterations for test case $TARGET without passing. Stopping for user input."
            exit 1
            ;;
        *)
            LVL_NUM=$(echo "$TARGET" | tr -dc '0-9')
            echo "Attempting automated iterative fix for $TARGET (max 3 iterations)..."
            i=1
            while [ $i -le 3 ]; do
                echo "--- Iteration $i / 3 for $TARGET ---"
                python test/level.py --level "$LVL_NUM"
                if [ $? -eq 0 ]; then
                    echo "Success! All tests in $TARGET passed on iteration $i."
                    exit 0
                fi
                echo "Tests failed. Running compiler build and re-evaluating..."
                python build.py
                i=$((i + 1))
            done
            echo "Reached maximum 3 iterations for $TARGET without passing all tests. Stopping for user input."
            exit 1
            ;;
    esac
else
    echo "Usage: sh run.sh [build|test [level1]|fix level1|fix T0104]"
    exit 1
fi
