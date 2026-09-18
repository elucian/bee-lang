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
    if [ -n "$TARGET" ] && [ "${TARGET#level}" = "$TARGET" ]; then
        # A bare test-case name (e.g. T0504), not a level: route to the single
        # test runner WITHOUT wiping test/status, test/output or .temp.
        echo "Running single test $TARGET..."
        python test/solo.py "$TARGET"
    else
        # A level (levelX) or no target: clean test status before running tests.
        python clean.py
        if [ -n "$TARGET" ]; then
            echo "Running test suite for $TARGET..."
            python test/test.py "$TARGET"
        else
            echo "Running complete test pipeline..."
            python test/test.py
        fi
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
elif [ "$COMMAND" = "sync" ]; then
    # Mirror the Bee tutorial between the local repo and the SCL live site.
    # Default is push (local -> SCL). Use `sh run.sh sync pull` to reverse.
    # Extra flags (--dry-run, --delete, etc.) are forwarded to the script.
    if [ "$TARGET" = "pull" ]; then
        python scripts/sync_tutorial.py --pull $REST
    elif [ "$TARGET" = "push" ]; then
        python scripts/sync_tutorial.py --push $REST
    else
        python scripts/sync_tutorial.py $TARGET $REST
    fi
elif [ "$COMMAND" = "smoke" ]; then
    echo "Running intelligent self-health check..."
    python test/smoke.py
elif [ "$COMMAND" = "ed" ]; then
    # bee-ed: native file-maintenance CLI for AI agents and humans.
    # Build on demand from the repo root, then forwards all remaining args
    # to the binary (which runs from the caller's current directory).
    ROOT="$(cd "$(dirname "$0")" && pwd)"
    CALLER="$(pwd)"
    cd "$ROOT"
    go build -o bin/bee-ed ./cmd/ed/ || { echo "bee-ed: build failed"; cd "$CALLER"; exit 1; }
    cd "$CALLER"
    # Forward the bee-ed args verbatim. The initial `shift 2` removed this
    # script's own (COMMAND, TARGET); TARGET still holds the bee-ed subcommand
    # and the remaining positionals are its flags. Each must stay one quoted
    # argv element, and MSYS2 path/arg conversion must be off, otherwise the
    # shell re-globs/converts patterns like `<code class` or `/`-containing
    # replacements (`</code></pre>`) and corrupts them before bee-ed sees them.
    MSYS2_ARG_CONV_EXCL='*' MSYS_NO_PATHCONV=1 "$ROOT/bin/bee-ed" "$TARGET" "$@"
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
        DECISIONS=$(git diff --cached -U0 -- manual/DECISIONS.md | grep '^+' | grep -oE 'Decision [0-9]+' | sort -u | tr '\n' ', ' | sed 's/, $//')
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
    echo "Usage: sh run.sh [build | test [level1] | check [level1] | solo <target> | reset [level1] | clean | smoke | sync [pull] [--dry-run] | ed <cmd...> | commit [message]]"
    exit 1
fi
