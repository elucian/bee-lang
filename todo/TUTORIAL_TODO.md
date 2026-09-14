# Bee Tutorial — Vision, Gaps & Improvement Todo

> The `/tutorial/` directory is the student-facing **design source-of-truth**.
> It must present every ratified `spec/` fact precisely, additively (no
> un-ratified features), and be complete enough that a student can build a
> correct Bee compiler from it alone.

## Guiding principles

1. **Spec is facts, tutorial is the complete design document.** A tutorial
   page that disagrees with `spec/` on a keyword, operator, or grammar
   production is a defect.
2. **Additive only.** Never introduce a feature that is not ratified (D10,
   D11, D8 remain out of scope until ratified).
3. **Locally edited, manually synced.** All edits land under `tutorial/`;
   `sh run.sh sync` is run by the user at end of day, never automatically.
4. **Highlighter deferred.** `tutorial/js/bee.js` is not touched until
   explicitly requested.
5. **Quiz removed.** Delete the certification/quiz artifacts; the tutorial is
   a design document, not an exam.

---

## Priority 0 — Blocking structural gaps

### 0.1 Create `tutorial/memory.html` (NEW FILE — highest priority)
- [x] **Done.** `tutorial/memory.html` created and aligned with
  `spec/00-memory-model.md` (three tiers, assignment semantics, `zap`/
  `E0801`, region cleanup, thread boundaries, diagnostic table, E08xx block; see §3.3).
- **Source:** `spec/00-memory-model.md`.
- **Why:** This is the only spec module with no dedicated page; today it is
  reduced to a bullet in `features.html`. The memory model is load-bearing for
  uniqueness (ARC + Regions + GC) and has no student-facing home.
- **Must cover:**
  - Three-tier model: ARC (mutable/heap) / Region Arena + `zap` (transient) /
    compacting GC (immutable strings & ropes).
  - Assignment semantics table: `:=` (assign / update-in-place), `::`
    (structural clone), `:` (structural binding — NOT mutation).
  - `new`, `let`, `alter` mutability modifiers and how `:=`/`::` differ per
    primitive vs boxed vs reference.
  - `zap identifier;` and the `E0801 AccessAfterZap` invariant + debug
    `UseAfterZap` panic.
  - Region cleanup triggers (`return`, `done`, `next`).
  - Thread boundaries: immutable sharing, ARC transfer (`LOCK XADD`), worker
    error isolation.
  - Diagnostic table `E0801`–`E0804` — memory block moved to `E08xx` on
    2026-09-14 to resolve the collision with `spec/04-structure.md` (see §3.3).

### 0.2 Drop the quiz / certification from `tutorial/index.html`
- [x] **Done.** Quiz/certification artifacts removed; index reorganized into the
  four-phase roadmap (Phase 1 Core, 2 Logic & Control, 3 Data & Collections,
  4 Advanced Systems) with per-phase topic numbering restarting at 1.
- Remove topic row `#15 … Certification` and the `#quiz` section + certificate
  alert + form link.
- Re-number the roadmap to reflect only real topics.
- Remove or repurpose `tutorial/prompt.txt` (currently quiz-generation
  leftovers).

### 0.3 Wire `memory.html` into the roadmap
- [x] Add `memory.html` to `index.html` as a topic — now Phase 1 topic #6.
- ~~Add a "Read next / previous" link.~~ Footer "Read next/previous/more"
  links were removed project-wide (navigation is sidebar-only now).
- [x] Add `tutorial/data/memory.json` (every other page has a sidebar/ToC
  `data/*.json`; `memory.html` is missing one, so it does not appear in the
  sidebar navigation).

---

## Priority 1 — Precision gaps in existing pages

Each item below is a spec fact that is missing, stale, or under-specified in
the matching tutorial page. Work item = reconcile tutorial with spec.

### 1.1 `functions.html`  ← `spec/07-functions.md`
- [x] Add formal **purity invariants** as a numbered, canonical list:
  statelessness (`no new/let/alter`), referential transparency, side-effect
  isolation (no I/O), and rule-isolation (rule→lambda OK; lambda→rule NOT).
- [x] Add `L` **type descriptor** section (first-class lambda values) and
  lambda type signatures `fn: λ(x,y∈R)=>R`.
- [x] Add **SIMD & GPU auto-vectorization** section (currently entirely
  missing) — pure lambdas over collections auto-vectorize (AVX-512) and can
  lower to OpenCL/GPU pipelines.
- [x] Add diagnostic table `E0701`–`E0705`.
- [x] Mark `x.type()` method-call dispatch as **debt** (T0126, issue 15) rather
  than implying it works today.

### 1.2 `types.html`  ← `spec/05-types.md`
- [x] **Done 2026-09-14.** Verified the full primitive catalogue (13 types:
  `B A U N Z R Q C S D T L G`) — all present and correct in the table.
- [x] **Done 2026-09-14.** Type promotion hierarchy added. The public site
  does **not** render mermaid, so the `spec/05` §5 DAG is presented as the
  canonical chain `B → N → Z → R → Q → C` plus the widening branches
  (`N→A`, `Z→U`), in a table with explicit-vs-implicit rules.
- [x] **Done 2026-09-14.** Subtypes (`<:`), range/domain notation
  (`..`, `..<`, `>..`, `>..<`), and domain step ratio present. **Corrected** the
  Domain Type section from the deprecated colon-step `(min..max:ratio)` to the
  canonical D13 postfix `(min..max)(step)` and repaired a `&\4` backslash
  mangling.
- [x] **Done.** Rational fixed-point `Q(m.n)` notation and `p\q` literal present.
- [x] **Done.** Approximate equality `≈` and tolerance `±` present.
- [x] **Done.** Explicit cast `:>` (incl. Float→Rational example) and `∈`/`in`
  type check present.
- [x] **Done.** Diagnostic table `E0501`–`E0506` present.
- [x] **Done.** `≡` confirmed as geometric congruence only — it never appears
  as value equality; the identity/value taxonomy lives in `operators.html`.

### 1.3 `rules.html`  ← `spec/03-rules.md`
- [x] **Done 2026-09-14.** Curried/named slots, assert/expect, forward decls,
  TCO, closures, result deconstruction, and the E03xx table were all already
  present. Added the missing **Companion & Singleton Rules** section (spec §5.1)
  and linked it from Advanced Topics.
- [x] Add **curried / named-parameter slots** (Decision 6): signature
  `rule f(a)(sep: ", " ∈ S)`, call site `f(a)(sep: "|")`, order
  independence, defaults, and deprecation of `using`.
- [x] Add `assert` vs `expect` contract semantics (warning vs fatal).
- [x] Add forward declarations (no hoisting) + mutual recursion.
- [x] Add tail-call optimization (TCO) note.
- [x] Add closures / state generators (boxed `[start]` capture).
- [x] Add companion & singleton rules.
- [x] Add result deconstruction `new s, d := f(...)`, wildcard `_`, and the
  multi-result-in-expression restriction.
- [x] Add diagnostic table (`E0301`–`E0309`, `W0301`, `W0308`, `E0010`, `E0011`).

### 1.4 `control.html`  ← `spec/02-statements.md` (+ D14/D15)
- [x] **Done 2026-09-14.** Add uniform `done [label];` terminator for all
  blocks (D14) — reinforced in the page intro with the full terminated-block
  list + `E0205 LabelMismatch` cross-ref.
- [x] **Done.** `next [label] [if cond]` canonical continue jump (D15) and
  `repeat` deprecated synonym (E0010) — in notes, nested/for cycles, and the
  diagnostic table.
- [x] **Done.** `stop`/`redo` semantics confirmed unchanged and documented.
- [x] **Done.** `start`/`with` scope blocks; `match` value-matching; `trial`
  error handling; `then` post-loop epilogue.
- [x] **Done 2026-09-14.** Add ternary/conditional expression selector `(expr_true
  if condition else expr_false)` — new section under Conditional Selector; parser
  + `test/level2/T0225.bee` already implement it.
- [x] **Done.** Statement diagnostic table (E0201–E0206, E0010, E0011).

### 1.5 `operators.html`  ← `spec/01-lexical-structure.md` §3 (+ D2/D7/D12/D13)
- [x] Enforce the identity/value taxonomy table: `=` value equality, `¬` value
  inequality, `is` / `is not` / `@a = @b` reference identity, `!` logical NOT.
  (verified 2026-09-14)
- [x] Confirm logical synonymy `and/or/xor/not` ↔ `∧/∨/⊕/¬` (D7) and that the
  lexer emits one token per pair.
- [x] Confirm `is not` is one token (Maximal-Munch).
- [x] Range operators `..`, `..<`, `>..`, `>..<` (D13); legacy `.!`, `!.`, `!!`
  deprecated.
- [x] `≡` reserved for **geometric congruence only**.
- [x] List deprecated forms (`≠`, `==`, `!=`, `<>`) with E0010 notes.
- [x] **D10 radicals (2026-09-14):** numeric-operators table fixed — `√` is a
  **prefix** radical (was wrongly written `x√n`), with degree forms `²√`/`³√`/
  `ⁿ√` and the rule that power inside the operand resolves first (`³√ 2³ = 2`).

### 1.6 `collections.html`  ← `spec/10-collections.md`
- [x] **Done 2026-09-14.** Verify **Ordinal** type `(start){ id, id, … }` — present (with `(n){…}` start specifier and public `.name` elements).
- [x] **Done.** Verify `$` end-anchor, 1-based indexing, and `$ - k` negative-relative — added a dedicated **1-Based Indexing & the `$` End Anchor** section (`list[$]`, `list[$ - 1]`, slicing `list[2..$ - 1]`, `E1006 ZeroBasedIndexAttempt`).
- [x] **Done.** Verify set algebra `∩ ∪ \ Δ ⊂ ⊃ Σ` and `∈` membership — added the missing `Σ` summation operator to the data-set example.
- [x] Verify matrix row-major + 2D indexing — already present.
- [x] Add diagnostic table `E1001`–`E1006` — already present and registry-aligned.

### 1.7 `processing.html`  ← `spec/11-processing.md`
- [x] **Done 2026-09-14.** Verify boxing `[x]` and unboxing `Type(boxed)` — present (&sect;Boxed values).
- [x] **Done 2026-09-14.** Quantifiers present; added ASCII synonyms `forall`/`exists` (D7 synonymy) to &sect;Logic Qualifiers.
- [x] **Done 2026-09-14.** Added &sect;Pipelines &amp; Map-Reduce: `>>` chaining, `.map/.filter/.reduce`, aggregates `.count/.sum/.avg/.min/.max`, and `E1103` cross-ref. Also corrected the stale &sect;List Operations claim that `>>` is a list-shift (it is the pipeline operator, spec/11 §4).
- [x] **Done.** Verify deconstruction `*tail` and matrix row/col slice `M[1,*]`, `M[*,2]` — present (&sect;Array decomposition, &sect;Matrix operations).
- [x] **Done.** Add diagnostic table `E1101`–`E1106` — present and registry-aligned.

### 1.8 `concurrency.html`  ← `spec/12-concurrency.md`
- [x] **Done 2026-09-14.** Verify `begin`/`wait`, thread-safe reduction `+>`,
  coroutines `yield`, channel extraction — present. Corrected the stale `wait`
  definition (it is a synchronization barrier, not a sleep timer), fixed the
  channel-extract glyph from `<-` to canonical `<<` (spec §4.2 / E1203),
  and replaced deprecated range `(1.!100:25)` with canonical D13
  `(1..<101)(25)`.
- [x] **Done 2026-09-14.** Added &sect;Worker Exception Isolation — `$trial`
  handle, barrier re-raise, `E1204` cross-ref.
- [x] **Done.** Add diagnostic table `E1201`–`E1205` — present and registry-aligned.

### 1.9 `graphics.html`  ← `spec/13-graphics.md`
- [x] **Done 2026-09-14.** Verify angular `G` type and `° ′ ″` literals — present.
- [x] **Done 2026-09-14.** Verify `CRT/POL/VEC/CRC/SQR/PLG` primitives and
  `Canvas/Layer/Shape/Label` — present. Corrected primitive signatures to the
  spec/13 §3 catalogue (POL `{r ∈ R,θ∈G}`, CRC `{o∈CRT,r∈R}`, SQR `{b∈R}`, PLG
  `{v∈[CRT]}`) and removed the un-ratified ARC/TRG/REG rows (additive violation).
- [x] **Done 2026-09-14.** Verify `draw/wipe/show/hide` and affine transforms —
  added a &sect;drawing-API code example and an &sect;Affine Transformations
  (rotate/translate/scale) section.
- [x] **Done 2026-09-14.** Added &sect;Geometric Congruence — `≡` congruence-only
  semantics (spec §5.1), never value equality.
- [x] **Done.** Add diagnostic table `E1301`–`E1305` — present and registry-aligned.

### 1.10 `library.html`  ← `spec/14-library.md`
- [x] **Done 2026-09-14.** Added the canonical <code>$bee.sys.io</code> namespace
  import (<code>use $bee.sys.io as IO;</code>) and <code>IO.File.open/write/close</code> +
  <code>IO.Folder.exist/create/list</code> API to &sect;File IO (replacing the
  un-ratified flat <code>File.open()</code> form); noted the <code>$bee.sys</code>
  sibling namespaces (math/string/time/env) per spec §1.
- [x] **Done.** Verify <code>entity.type()</code>, <code>size</code>, <code>length</code>,
  <code>capacity</code> — present as introspection rules.
- [x] **Done 2026-09-14.** Verified <code>$error</code> allocation — updated &sect;Error
  Type to the spec/14 §5 ranges (<code>1..199</code> system, <code>200..9999</code> user,
  <code>≤ -1</code> panic) and tied it to the <code>$error</code> system variable.
- [x] **Done.** Add diagnostic table `E1401`–`E1405` — present and registry-aligned.

### 1.11 `structure.html`  ← `spec/04-structure.md`
- [x] **Done 2026-09-14.** Verify `$` sigil system-variable table — added
  &sect;System Path Sigils table (<code>$bee_home $bee_lib $pro_home $pro_lib
  $pro_mod $pro_log $pro_src</code>) per spec/04 §3.
- [x] **Done.** Verify export prefix `.`, <code>with … do … done</code>, <code>alias</code> —
  present (&sect;Name space, &sect;Global scope).
- [x] **Done.** Add diagnostic table (`E0401`–`E0406`) — present, includes the
  spec/04 §7 <code>E0406 UnsynchronizedWorker</code> row, registry-aligned.

### 1.12 `syntax.html`  ← `spec/01-lexical-structure.md`
- [x] **Done.** Verify UTF-8/rune handling description, BOM rejection (`E0101`) —
  BOM covered via `E0101 UTF8BomMarker` row in the diagnostic table.
- [x] **Done.** Verify identifier casing rules — present (&sect;Identifiers).
  Reserved single-letter type list conflict resolved in §3.2 (13 canonical
  types `B A U N Z R Q C S D T L G` in both spec/01 and spec/05).
- [x] **Done 2026-09-14.** Verify comment forms — added the missing
  <code>(: ... :)</code> expression comment to &sect;Comments (spec §2.1.3);
  <code>--</code> and <code>+- -+</code> already present.
- [x] **Done 2026-09-14.** Added &sect;String Literals &amp; Markup — the
  <code>'…'</code> / <code>"…"</code> interpolated <code>#(...)</code> / backtick
  raw forms, escape-sequence list, and markup blocks
  (<code>&lt;text&gt; &lt;sql&gt; &lt;html&gt; &lt;xml&gt; &lt;json&gt; &lt;code&gt;</code>).
- [x] **Done.** Add lexical diagnostic table `E0101`–`E0105` — present and
  registry-aligned (descriptor references spec/01 §6).

---

## Priority 2 — New pages to make it a *complete* design document

These are additive, high-value pages for the "hundreds of students building
compilers" audience. Treat as proposals; confirm scope before authoring.

### 2.1 `diagnostics.html` (NEW)
- [x] **Done 2026-09-14.** Consolidated, cross-referenced table of **every**
  diagnostic code (`E00xx` global + `E01xx`→`E14xx`, `Wxx`) grouped by module,
  with condition + resolution.
- [x] **Done.** Generated directly from `registry/diagnostics.json` via
  `scripts/gen_diagnostics_page.py` (single source of truth; re-runnable), wired
  into the `index.html` roadmap as Phase 4 topic #4, with sidebar ToC
  `data/diagnostics.json`.
- [x] E04xx collision resolved (§3.3); page documents the `E08xx` memory block
  and the collision policy.

### 2.2 `grammar.html` (NEW, optional)
- Consolidated EBNF grammar reference drawn from every spec module.
- Value: a single machine/mind-readable grammar the student can implement from.

### 2.3 `roadmap.html` / design-rationale (NEW, optional)
- A narrative of *why* Bee's design choices exist (rule-vs-function split,
  1-based indexing rationale, memory tiers) — currently scattered across pages.

---

## Priority 3 — Cross-cutting data-quality / consistency items

These are prerequisites or supporting chores, not page authoring.

### 3.1 `todo/DECISIONS.md` index is stale
- [x] **Done 2026-09-14.** D10 and D11 ratified; the body and index now agree
  (`✅ Ratified` in both) and `MANIFEST.md` mirrors the ratification. D10's
  MANIFEST entry was also corrected to drop the stale `¬`-as-unary claim (that
  belongs to D12) and D11's summary captured the accepted parenthesised form
  `new (a, b):(1,2) ∈ Z` and the rejected bare list. Index also includes D15.

### 3.2 Single-letter built-in type list is inconsistent
- [x] **Resolved.** `spec/01` now lists the canonical 13 reserved single-letter
  types `B A U N Z R Q C S D T L G` (with `A`=Alpha, `L`=Lambda, `G`=Angular),
  matching the `spec/05` catalogue and the `types.html` table exactly.
- `Array`, `Map`, `List`, `Graph` are now explicitly named (non-letter) complex
  types, identified by the `Type` column rather than a reserved letter.

### 3.3 `E04xx` diagnostic block was overloaded — **RESOLVED 2026-09-14**
- [x] Created `registry/diagnostics.json` as the single source of truth for all
  `E`/`W` codes (per user directive).
- [x] Bumped and re-registered `spec/00` (memory) from the colliding `E0401`–`E0404`
  to the free `E08xx` block (`E0801`–`E0804`). `spec/04` retains the canonical
  `E04xx` block. Updated `spec/00-memory-model.md` and `tutorial/memory.html`;
  removed the "E04xx overloaded" warning alert. See `registry/README.md`.
- [x] **Completed 2026-09-14:** All per-module diagnostic tables are present and
  section-referenced. Verified every `<em>…</em> module (<code>spec/NN</code> §N)`
  intro against the registry `source` fields; corrected `operators.html` (§3→§6),
  `operators.html` (§3→§6),
  `syntax.html` (§1→§6, + stripped stray `</p>` from the table header row), and
  `rules.html` (no § → §7). Added the missing `<em>Functions</em> module`
  descriptor paragraph in `functions.html`. Repaired one mangled `E0102` row in
  `operators.html` and an orphaned `<p>` in `structure.html`.
- [x] **2026-09-14 — HTML tag-balance repair (1.7–1.12 pass):** Fixed the
  recurring corruption that breaks tag nesting — the mangled
  `<</p>!-- Footer -->` footers in `processing.html`, `structure.html`, and
  `objects.html` (restored to `<!-- Footer -->`); a missing `</li>` in
  `structure.html` (§Secondary modules); a reversed `</pre></code>` close-order
  in `structure.html` (§Name space aliases); and unclosed `<p>` tags in
  `processing.html` (§THE END), `objects.html` (§Array of Objects), and
  `syntax.html` (§Superscript). All seven tutorial pages now balance to the
  wrapper under an HTML5 parser (verified via Python `html.parser`).

### 3.4 `features.html` overloaded + link drift
- [x] `features.html` no longer carries memory-model notes (verified 2026-09-14:
  grep for memory/zap/ARC/three-tier is empty), so the trim-to-`memory.html`
  subtask is moot. Features now cross-links via the roadmap/index.
- [x] ~~Audit/remove the "Read next"/"previous" footer links.~~ Done: all
  `Read next/previous/more` footer blocks removed from `tutorial/*.html`.

### 3.5 Confirm source-of-truth framing
- **Discovery, not documents.** Neither `/spec` nor `/tutorial` is the source of
  truth. The language truth is *discovered* through a design loop: a feature is
  proposed from needs / other languages, then adapted, tuned, simplified, and
  combined to stay consistent with the rest of the design (trial and error
  happens here). Only once its shape is stable do we express it in **both** the
  tutorial (usecases + rationale for students) and the spec (formal facts) as
  twin outputs of the same decision.
- **Implication for agents:** both files are *snapshots of the current design*,
  not immutable axioms. They must agree with each other and remain open to
  revision as discovery continues. When a design changes, update the matched
  spec ↦ tutorial pair together (see §6 mapping and the sync invariant).
- **DoD note:** using a feature in the tutorial is not ratification; ratification
  is the design decision itself, recorded in `todo/DECISIONS.md` (D1–D16).

---

## Definition of done (per page)

A page is "precise" when:
1. Every keyword/operator/grammar production it shows matches a ratified
   `spec/` fact verbatim in intent.
2. It cites decision ids (D1, D6, D7, D12, D13, D14, D15) where a choice rides
   on a ratified decision.
3. It includes the module's diagnostic-code table, with every code first
   registered in `registry/diagnostics.json`.
4. It uses canonical operator forms (`¬`, `is`, `is not`, `=` for value
   equality, `≡` only for geometry) — no deprecated glyphs except in an
   explicit deprecation note.
5. It links forward/backward consistently in the roadmap.

## Execution order (suggested)

1. 0.1–0.3 (memory.html + remove quiz + rewire roadmap) — the structural win.
2. 1.1–1.12 (precision passes), spec-to-tutorial.
3. 3.1–3.5 (data-quality fixes) in parallel where they gate a page.
4. 2.1–2.3 (new reference pages) after the core is stable.
5. Final local `sh run.sh sync` (manual, user-triggered).
