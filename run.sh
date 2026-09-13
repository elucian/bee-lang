#!/bin/sh
# run.sh - Master CLI helper script for Bee compiler workflow

COMMAND="$1"
TARGET="$2"
shift 2 2>/dev/null || shift $#
REST="$@"

if [ "$COMMAND" = "build" ]; then
    python clean.py
    python build.py
elif [ "$COMMAND" = "clean" ] || [ "$COMMAND" = "clan" ]; then
    echo "Cleaning test outputs and telemetry..."
    python clean.py
elif [ "$COMMAND" = "test" ]; then
    # Clean test status before running tests
    python clean.py
    
    if [ -n "$TARGET" ]; then
        echo "Running test suite for $TARGET..."
        python test/test.py "$TARGET"
    else
        echo "Running complete test pipeline..."
        python test/test.py
    fi
elif [ "$COMMAND" = "check" ]; then
    if [ -n "$TARGET" ]; then
        echo "Checking syntax for $TARGET..."
        python scripts/check.py "test/$TARGET"
    else
        echo "Checking all tests syntax..."
        python scripts/check.py "test"
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
elif [ "$COMMAND" = "smoke" ]; then
    echo "Running intelligent self-health check..."
    python test/smoke.py
elif [ "$COMMAND" = "commit" ]; then
    # Stage everything, synthesize a message from the diff (or use args), commit, push.
    git add -A
    if git diff --cached --quiet; then
        echo "Nothing to commit."
        exit 0
    fi
    MSG=$(echo "$TARGET $REST" | sed 's/^ *//;s/ *$//')
    if [ -z "$MSG" ]; then
        # Synthesize subject: detect ratified decisions and new tests in staged diff.
        DECISIONS=$(git diff --cached -U0 -- todo/DECISIONS.md | grep '^+' | grep -oE 'Decision [0-9]+' | sort -u | tr '\n' ', ' | sed 's/, $//')
        NEWTESTS=$(git diff --cached --name-only --diff-filter=A -- 'test/level*/T*.bee' | grep -oE 'T[0-9]+' | sort -u | tr '\n' ' ' | sed 's/ $//')
        FILES=$(git diff --cached --name-only | wc -l | tr -d ' ')
        if [ -n "$DECISIONS" ] && [ -n "$NEWTESTS" ]; then
            MSG="Ratify $DECISIONS; add tests $NEWTESTS"
        elif [ -n "$DECISIONS" ]; then
            MSG="Ratify $DECISIONS"
        elif [ -n "$NEWTESTS" ]; then
            MSG="Add tests $NEWTESTS"
        else
            AREAS=$(git diff --cached --name-only | cut -d/ -f1 | sort -u | tr '\n' ', ' | sed 's/, $//')
            MSG="Update $FILES files across $AREAS"
        fi
        echo "Synthesized message: $MSG"
    fi
    GIT_EDITOR=true git commit -m "$MSG"
    git push
else
    echo "Usage: sh run.sh [build | test [level1] | check [level1] | solo <target> | reset [level1] | clean | smoke | commit [message]]"
    exit 1
fi
