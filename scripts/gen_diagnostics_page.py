"""Generate tutorial/diagnostics.html and tutorial/data/diagnostics.json
directly from registry/diagnostics.json (the single source of truth).

Usage: python scripts/gen_diagnostics_page.py
"""
import json
import os

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
REG = os.path.join(ROOT, "registry", "diagnostics.json")
OUT_HTML = os.path.join(ROOT, "tutorial", "diagnostics.html")
OUT_JSON = os.path.join(ROOT, "tutorial", "data", "diagnostics.json")

MODULES = [
    ("00-memory-model", "E00 / E08", "Memory model",
     "<em>Memory</em> module <code>spec/00-memory-model.md</code> &sect;5. "
     "Block moved from the overloaded <code>E04xx</code> range to the free "
     "<code>E08xx</code> on 2026-09-14 (spec/04 retains "
     "<code>E0401</code>&ndash;<code>E0406</code>). Related page: "
     "<code>memory.html</code>."),
    ("01-lexical-structure", "E01", "Lexical structure",
     "<em>Lexical structure</em> module <code>spec/01-lexical-structure.md</code> "
     "&sect;6. Related page: <code>syntax.html</code>."),
    ("02-statements", "E02", "Statements",
     "<em>Statements</em> module <code>spec/02-statements.md</code> &sect;7 "
     "(+ D14/D15). Related page: <code>control.html</code>."),
    ("03-rules", "E03", "Rules",
     "<em>Rules</em> module <code>spec/03-rules.md</code> &sect;7 (+ D6). "
     "Related page: <code>rules.html</code>."),
    ("04-structure", "E04", "Structure",
     "<em>Structure</em> module <code>spec/04-structure.md</code> &sect;7. "
     "Retains the canonical <code>E04xx</code> block after memory was bumped to "
     "<code>E08xx</code>. Related page: <code>structure.html</code>."),
    ("05-types", "E05", "Types",
     "<em>Types</em> module <code>spec/05-types.md</code> &sect;7. "
     "Related page: <code>types.html</code>."),
    ("06-objects", "E06", "Objects",
     "<em>Objects</em> module <code>spec/06-objects.md</code> &sect;7. "
     "Related page: <code>objects.html</code>."),
    ("07-functions", "E07", "Functions",
     "<em>Functions</em> module <code>spec/07-functions.md</code> &sect;7. "
     "Related page: <code>functions.html</code>."),
    ("10-collections", "E10", "Collections",
     "<em>Collections</em> module <code>spec/10-collections.md</code> &sect;5. "
     "Related page: <code>collections.html</code>."),
    ("11-processing", "E11", "Processing",
     "<em>Processing</em> module <code>spec/11-processing.md</code> &sect;7. "
     "Related page: <code>processing.html</code>."),
    ("12-concurrency", "E12", "Concurrency",
     "<em>Concurrency</em> module <code>spec/12-concurrency.md</code> &sect;7. "
     "Related page: <code>concurrency.html</code>."),
    ("13-graphics", "E13", "Graphics",
     "<em>Graphics</em> module <code>spec/13-graphics.md</code> &sect;7. "
     "Related page: <code>graphics.html</code>."),
    ("14-library", "E14", "Library",
     "<em>Library</em> module <code>spec/14-library.md</code> &sect;7. "
     "Related page: <code>library.html</code>."),
]

HEAD = """<!DOCTYPE html>
<html lang="en" data-bs-theme="dark"><head>
  <meta charset="utf-8">
  <meta name="description" content="Bee diagnostics reference: every E/W diagnostic code grouped by module, with condition and resolution.">
  <meta name="author" content="Elucian Moise">
  <meta name="keywords" content="sage, code, bee, bee-lang, diagnostics, errors, warnings, compiler, reference">
  <meta name="viewport" content="width=device-width, initial-scale=1.0, viewport-fit=cover">
  <title>Bee Diagnostics Reference</title>
  <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet" crossorigin="anonymous">
  <link rel="icon" type="image/png" href="/images/favicon.ico">
  <script src="/projects/bee/js/bee.js"></script>
  <link rel="stylesheet" href="/assets/css/sage-common.css">  <link rel="stylesheet" href="/assets/css/content-topic.css">    <link rel="stylesheet" href="/assets/css/content-code.css">    <link rel="stylesheet" href="/assets/css/content-sidebar.css">
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bootstrap-icons@1.11.3/font/bootstrap-icons.min.css">
</head>
<body data-topic-runtime-injected="true" onload="bee_render()">
<div class="container">
  <header id="dynamic-header" class="container-fluid pb-2"></header>
  <div class="container-fluid px-0">
    <div class="row g-0">
      <aside class="side-bar col-lg-3 col-12">
        <div id="study-sidebar" class="sidebar-content shadow-sm p-3 sticky-top">
          <div class="d-flex justify-content-between align-items-center mb-2">
            <h5 class="mb-0">Lab Topics</h5>
          </div>
          <hr>
          <ul id="bookmark-list" class="list-unstyled">
          </ul>
        </div>
      </aside>
      <main id="main-content" class="col-lg-9 col-12 order-2 order-lg-1 p-3">
<div class="container">
"""

MID = """
<h1 id="diagnostics-reference">Diagnostics Reference</h1>
<div class="alert alert-secondary shadow-sm">
  Every Bee diagnostic code, grouped by module. This is a compiler-builder's single
  source of truth for the error surface: each code is first registered in
  <code>registry/diagnostics.json</code> before it may appear in any
  <code>/spec</code> or <code>/tutorial</code> table. The tables below are generated
  faithfully to that registry.
</div>

<h2 id="legend">Severity legend</h2>
<table class="table table-bordered table-striped table-sm">
  <thead>
    <tr><th>Severity</th><th>Meaning</th></tr>
  </thead>
  <tbody>
    <tr><td><code>E</code></td><td>Hard error &mdash; fatal; halts the build or traps at runtime unless handled.</td></tr>
    <tr><td><code>W</code></td><td>Soft warning &mdash; non-fatal; logged to <code>stderr</code> and does not change the exit status.</td></tr>
  </tbody>
</table>

<h2 id="conventions">Coding convention</h2>
<ul>
  <li><b>Block rule:</b> module <code>&lt;dd&gt;</code> owns block <code>E&lt;dd&gt;xx</code>, where <code>&lt;dd&gt;</code> is the module's leading digit (<code>01</code> &rarr; <code>E01</code>, <code>02</code> &rarr; <code>E02</code>, &hellip; <code>14</code> &rarr; <code>E14</code>).</li>
  <li><b>Memory block:</b> module <code>00</code> (memory model) has no natural <code>E00xx</code> range (<code>E00xx</code> is reserved for global/deprecation codes), so memory uses the first free block <code>E08xx</code>.</li>
  <li><b>Collision policy:</b> if a proposed code already exists with different semantics, the newer use is bumped to the next free slot in its module block and re-registered. A skipped slot within a block is therefore intentional.</li>
</ul>

<h2 id="global">Global codes</h2>
<p>Cross-cutting codes shared by every module. Up-scope targets for Phase 7.2
deprecation hardening. Source: <code>spec/00,01,02,03</code> + decisions D12/D14/D15.</p>
"""

TABLE_HEAD = """<table class="table table-bordered table-striped table-sm">
  <thead>
    <tr><th>Code</th><th>Severity</th><th>Condition</th><th>Resolution</th></tr>
  </thead>
  <tbody>
"""

TABLE_FOOT = """  </tbody>
</table>
"""

def esc(text):
    return text.replace("&", "&amp;").replace("<", "&lt;").replace(">", "&gt;")


def build_rows(codes):
    out = []
    for code in codes:
        c = codes[code]
        cond = esc(c["condition"])
        res = esc(c["resolution"])
        out.append(
            "    <tr><td><code>{}</code></td><td><code>{}</code></td>"
            "<td>{}</td><td>{}</td></tr>".format(code, c["severity"], cond, res)
        )
    return "\n".join(out)


def build_section(anchor, heading, intro, codes):
    block_rows = build_rows(codes)
    return (
        "\n<h2 id=\"{a}\">{h}</h2>\n<p>{i}</p>\n{t}{r}\n  </tbody></table>\n"
        .format(a=anchor, h=heading, i=intro, t=TABLE_HEAD, r=block_rows)
    )


def main():
    with open(REG, encoding="utf-8") as fh:
        reg = json.load(fh)
    modules = reg["modules"]

    body = HEAD + MID

    # global codes table
    gcodes = modules["global"]["codes"]
    body += TABLE_HEAD + build_rows(gcodes) + TABLE_FOOT

    for key, band, title, intro in MODULES:
        codes = modules[key]["codes"]
        anchor = {
            "00-memory-model": "e00-memory",
            "01-lexical-structure": "e01-lexical",
            "02-statements": "e02-statements",
            "03-rules": "e03-rules",
            "04-structure": "e04-structure",
            "05-types": "e05-types",
            "06-objects": "e06-objects",
            "07-functions": "e07-functions",
            "10-collections": "e10-collections",
            "11-processing": "e11-processing",
            "12-concurrency": "e12-concurrency",
            "13-graphics": "e13-graphics",
            "14-library": "e14-library",
        }[key]
        heading = "{} — {}".format(band, title)
        body += build_section(anchor, heading, intro, codes)

    body += """
<div class="alert alert-success shadow-sm">
  <b>Registry contract:</b> when a diagnostic is added, register it first in
  <code>registry/diagnostics.json</code> in the free slot of its module block. If it
  collides with an existing code, bump the newer use to the next free slot and
  re-register &mdash; never silently reuse an occupied code with different semantics.
  Keep this page and every per-module diagnostic table in mirror agreement.
</div>
"""

    body += """</div>
      </main>
    </div>
  </div>
  <hr>
  <footer class="footer copyright">
    <p class="x-small text-secondary mb-0">&copy; 2026 Sage-Code Laboratory</p>
  </footer>
</div>
<button id="open-sidebar" class="btn btn-primary d-lg-none shadow-lg" type="button">
  <span style="font-size: 24px;">&#9776;</span>
</button>
<script>
  window.TOPIC_CONFIG = {
    labId: 'bee',
    topicId: 'diagnostics',
    homeLink: '/projects/bee/#topics',
    labHomeLink: '/projects/bee/',
    inlineContent: true
  };
</script>
<script src="/assets/js/sage.js" defer></script>
<script src="/assets/js/topic-loader.js" defer></script>
</body>
</html>
"""

    with open(OUT_HTML, "w", encoding="utf-8") as fh:
        fh.write(body)

    # sidebar ToC json
    toc = [{
        "title": "Diagnostics Reference",
        "link": "#diagnostics-reference",
        "children": [
            {"title": "Severity legend", "link": "#legend"},
            {"title": "Coding convention", "link": "#conventions"},
            {"title": "Global codes", "link": "#global"},
            {"title": "E00 / E08 — Memory model", "link": "#e00-memory"},
            {"title": "E01 — Lexical structure", "link": "#e01-lexical"},
            {"title": "E02 — Statements", "link": "#e02-statements"},
            {"title": "E03 — Rules", "link": "#e03-rules"},
            {"title": "E04 — Structure", "link": "#e04-structure"},
            {"title": "E05 — Types", "link": "#e05-types"},
            {"title": "E06 — Objects", "link": "#e06-objects"},
            {"title": "E07 — Functions", "link": "#e07-functions"},
            {"title": "E10 — Collections", "link": "#e10-collections"},
            {"title": "E11 — Processing", "link": "#e11-processing"},
            {"title": "E12 — Concurrency", "link": "#e12-concurrency"},
            {"title": "E13 — Graphics", "link": "#e13-graphics"},
            {"title": "E14 — Library", "link": "#e14-library"},
        ],
    }]
    with open(OUT_JSON, "w", encoding="utf-8") as fh:
        json.dump(toc, fh, ensure_ascii=False, indent=2)

    print("wrote", OUT_HTML)
    print("wrote", OUT_JSON)


if __name__ == "__main__":
    main()
