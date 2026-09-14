"""Sync the Bee tutorial between the local repository and the SCL live site.

The Bee tutorial is the *source of truth* for the rendered HTML pages served
on the public SCL site. It is edited here, under `bee-lang/tutorial/`, and
mirrored into the external SCL repository directory that the user owns and
pushes to the live site independently.

This script is a plain-Python mirror (the environment has no `rsync`) that
copies files safely in one direction at a time:

    # Push local -> SCL (default): publish your edits to the live site.
    python scripts/sync_tutorial.py            # or: sh run.sh sync
    python scripts/sync_tutorial.py --dry-run
    python scripts/sync_tutorial.py --delete   # also remove SCL files not
                                               # present locally (true mirror)

    # Pull SCL -> local: adopt changes made over in the SCL repository.
    python scripts/sync_tutorial.py --pull
    python scripts/sync_tutorial.py --pull --dry-run

Sources/destinations are resolved from (in order of precedence):
  1. The --local and --scl command-line arguments.
  2. The SCL_BEE_DIR and BEE_TUTORIAL_DIR environment variables.
  3. Built-in defaults (project-relative local dir; known SCL path).
"""

import argparse
import os
import shutil
import sys

# Defaults. The SCL path is a Windows absolute path; the local path is
# resolved relative to this script's location so it works from any CWD.
_SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
_LOCAL_DEFAULT = os.path.normpath(os.path.join(_SCRIPT_DIR, "..", "tutorial"))
_SCL_DEFAULT = r"C:\Users\eluci\sage-code\scl\projects\bee"


def _abspath(p: str) -> str:
    return os.path.normpath(os.path.abspath(os.path.expanduser(p)))


def resolve_paths(args) -> tuple[str, str]:
    """Return (push_source, push_dest) for the requested direction.

    For a push, source is local and dest is SCL; for a pull they swap.
    """
    if args.local:
        local = _abspath(args.local)
    else:
        local = _abspath(os.environ.get("BEE_TUTORIAL_DIR", _LOCAL_DEFAULT))

    if args.scl:
        scl = _abspath(args.scl)
    else:
        scl = _abspath(os.environ.get("SCL_BEE_DIR", _SCL_DEFAULT))

    if args.pull:
        return scl, local
    return local, scl


def _needs_copy(src_file: str, dst_file: str) -> bool:
    """Return True if dst is missing or differs from src (size or mtime)."""
    if not os.path.exists(dst_file):
        return True
    s_st = os.stat(src_file)
    d_st = os.stat(dst_file)
    if s_st.st_size != d_st.st_size:
        return True
    # Compare modification time at nanosecond resolution when available.
    if getattr(s_st, "st_mtime_ns", None) and getattr(d_st, "st_mtime_ns", None):
        return s_st.st_mtime_ns != d_st.st_mtime_ns
    return s_st.st_mtime != d_st.st_mtime


def mirror(source: str, dest: str, delete_extras: bool, dry_run: bool):
    """Mirror `source` into `dest`.

    Returns a dict of counters describing what was (or would be) done.
    """
    if not os.path.isdir(source):
        print(f"error: source directory not found: {source}", file=sys.stderr)
        sys.exit(1)

    counters = {"copied": 0, "skipped": 0, "deleted": 0}
    source_set = set()  # normalized relative paths present in source

    for root, _dirs, files in os.walk(source):
        rel_root = os.path.relpath(root, source)
        for name in files:
            src_file = os.path.join(root, name)
            # rel path uses forward slashes regardless of platform
            rel_path = os.path.join(rel_root, name).replace(os.sep, "/")
            source_set.add(rel_path)

            dst_file = os.path.join(dest, rel_root, name)
            if _needs_copy(src_file, dst_file):
                if dry_run:
                    print(f"  copy  {rel_path}")
                else:
                    os.makedirs(os.path.dirname(dst_file), exist_ok=True)
                    shutil.copy2(src_file, dst_file)
                counters["copied"] += 1
            else:
                counters["skipped"] += 1

    if delete_extras and os.path.isdir(dest):
        for root, dirs, files in os.walk(dest, topdown=False):
            rel_root = os.path.relpath(root, dest)
            for name in files:
                rel_path = os.path.join(rel_root, name).replace(os.sep, "/")
                if rel_path not in source_set:
                    target = os.path.join(root, name)
                    if dry_run:
                        print(f"  rm    {rel_path}")
                    else:
                        os.remove(target)
                    counters["deleted"] += 1
            # Remove now-empty directories (never the dest root itself).
            if rel_root != "." and not os.listdir(root):
                if not dry_run:
                    os.rmdir(root)

    return counters


def main(argv=None) -> int:
    parser = argparse.ArgumentParser(
        description="Mirror the Bee tutorial between local and SCL repositories.",
    )
    parser.add_argument(
        "--push",
        action="store_true",
        help="push local -> SCL (default direction)",
    )
    parser.add_argument(
        "--pull",
        action="store_true",
        help="pull SCL -> local",
    )
    parser.add_argument(
        "--delete",
        action="store_true",
        help="delete destination files not present in the source (true mirror)",
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="print planned operations without modifying anything",
    )
    parser.add_argument(
        "--local",
        metavar="PATH",
        help="local tutorial directory (default: <repo>/tutorial)",
    )
    parser.add_argument(
        "--scl",
        metavar="PATH",
        help="external SCL bee directory (default: %s)" % _SCL_DEFAULT,
    )
    args = parser.parse_args(argv)

    if args.push and args.pull:
        print("error: choose at most one of --push / --pull", file=sys.stderr)
        return 2
    # Default is push.
    if not args.pull:
        args.pull = False

    source, dest = resolve_paths(args)
    direction = "pull (SCL -> local)" if args.pull else "push (local -> SCL)"

    print(f"Sync tutorial  {direction}")
    print(f"  source: {source}")
    print(f"  dest:   {dest}")
    if args.dry_run:
        print("  mode:   dry-run (no changes)")

    counters = mirror(source, dest, args.delete, args.dry_run)

    verb = "Would copy" if args.dry_run else "Copied"
    deleted_note = ""
    if args.delete:
        deleted_note = f", {'would remove' if args.dry_run else 'removed'} {counters['deleted']}"
    print(
        f"{verb} {counters['copied']} file(s), skipped {counters['skipped']}"
        f" unchanged{deleted_note}."
    )
    return 0


if __name__ == "__main__":
    sys.exit(main())
