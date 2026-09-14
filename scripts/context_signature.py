#!/usr/bin/env python3
"""Compute the Bee-Lang **context signature**.

The context signature is a single, stable, deterministic string that uniquely
identifies the current "context bundle" — the normative design facts that an
LLM agent must hold to work on Bee correctly. It is meant to be prepended to
every task prompt so a prompt cache can recognize an unchanged context and
reuse it, increasing the cache-hit rate and lowering the cost of re-building
context from scratch.

Design guarantees
-----------------
- **Deterministic:** the signature is a pure function of a fixed, ordered set
  of files. No timestamps, no environment, no randomness.
- **Minimal & complete:** it changes if and only if a *normative fact*
  changes (a ratified decision, the manifest phase/version, or any spec
  module), ensuring an exact correspondence between signature and cacheable
  context.
- **Human-readable prefix:** `BEE-LANG-V<phase>-D<epoch>-<hash8>` lets a human
  or agent recognize the snapshot at a glance while the hash encodes exact
  content.

Output format
-------------
    BEE-LANG-V8.7-D15-1a2b3c4d

Components
----------
- `V<phase>`   : current Manifest Phase/version, read from MANIFEST.md.
- `D<epoch>`   : highest authored decision id in todo/DECISIONS.md.
- `<hash8>`    : first 8 hex chars of SHA-256 over the canonical files below.

Usage
-----
    python scripts/context_signature.py
"""

import hashlib
import re
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent

# Canonical context files, in a FIXED order (do not reorder — it changes the hash).
FILES = [
    "GEMINI.md",
    "MANIFEST.md",
    "todo/DECISIONS.md",
]
# All spec modules, sorted, excluding the index (readme.md).
FILES += sorted(
    str(p.relative_to(ROOT))
    for p in (ROOT / "spec").glob("*.md")
    if p.name != "readme.md"
)


def _read(name: str) -> str:
    return (ROOT / name).read_text(encoding="utf-8")


def current_phase() -> str:
    """Extract the Manifest Phase version (e.g. '8.7')."""
    m = _read("MANIFEST.md")
    mm = re.search(r"Phase\s+(\d+)\.(\d+)", m)
    if mm:
        return f"{mm.group(1)}.{mm.group(2)}"
    return "0.0"


def decision_epoch() -> int:
    """Highest authored decision id referenced in the decisions backlog."""
    d = _read("todo/DECISIONS.md")
    ids = sorted({int(mm.group(1)) for mm in re.finditer(r"\bD(\d+)\b", d)})
    return ids[-1] if ids else 0


def fingerprint() -> str:
    """SHA-256 over the canonical files, concatenated with path delimiters."""
    h = hashlib.sha256()
    for name in FILES:
        h.update(("<<<" + name + ">>>\n").encode("utf-8"))
        h.update(_read(name).encode("utf-8"))
        h.update(b"\n")
    return h.hexdigest()


def main() -> int:
    phase = current_phase()
    epoch = decision_epoch()
    digest = fingerprint()
    print(f"BEE-LANG-V{phase}-D{epoch}-{digest[:8]}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
