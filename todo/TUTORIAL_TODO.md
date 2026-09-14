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
- [ ] Verify the full primitive catalogue (13 types: `B A U N Z R Q C S D T L G`)
  is present and correct; add any missing (`D` Date, `T` Time, `G` Angular).
- [ ] Add promotion-hierarchy diagram (mermaid) from `spec/05` §5.
- [ ] Add subtypes (`<:`), range/domain notation (`..`, `..<`, `>..`, `>..<`),
  and domain step ratio — aligned to D13.
- [ ] Add rational fixed-point `Q(m.n)` notation and `p\q` literal.
- [ ] Add approximate equality `≈` and tolerance `±`.
- [ ] Add explicit cast `:>` (Float→Rational example) and `∈`/`in` type check.
- [ ] Add diagnostic table `E0501`–`E0506`.
- [ ] Confirm `≡` appears **only** as geometric congruence (Phase 8.7),
  never as value equality.

### 1.3 `rules.html`  ← `spec/03-rules.md`
- [ ] Add **curried / named-parameter slots** (Decision 6): signature
  `rule f(a)(sep: ", " ∈ Str)`, call site `f(a)(sep: "|")`, order
  independence, defaults, and deprecation of `using`.
- [ ] Add `assert` vs `expect` contract semantics (warning vs fatal).
- [ ] Add forward declarations (no hoisting) + mutual recursion.
- [ ] Add tail-call optimization (TCO) note.
- [ ] Add closures / state generators (boxed `[start]` capture).
- [ ] Add companion & singleton rules.
- [ ] Add result deconstruction `new s, d := f(...)`, wildcard `_`, and the
  multi-result-in-expression restriction.
- [ ] Add diagnostic table (`E0301`–`E0309`, `W0301`, `W0308`, `E0010`, `E0011`).

### 1.4 `control.html`  ← `spec/02-statements.md` (+ D14/D15)
- [ ] Add uniform `done [label];` terminator for all blocks (D14).
- [ ] Add `next [label] [if cond]` canonical continue jump (D15) and mark
  `repeat` as deprecated synonym (E0010).
- [ ] Confirm `stop`/`redo` semantics unchanged and documented.
- [ ] Add `start`/`with` scope blocks; `match` value-matching; `trial` error
  handling; `then` post-loop epilogue.
- [ ] Add ternary/conditional selector syntax.
- [ ] Add statement diagnostic table.

### 1.5 `operators.html`  ← `spec/01-lexical-structure.md` §3 (+ D2/D7/D12/D13)
- [ ] Enforce the identity/value taxonomy table: `=` value equality, `¬` value
  inequality, `is` / `is not` / `@a = @b` reference identity, `!` logical NOT.
- [ ] Confirm logical synonymy `and/or/xor/not` ↔ `∧/∨/⊕/¬` (D7) and that the
  lexer emits one token per pair.
- [ ] Confirm `is not` is one token (Maximal-Munch).
- [ ] Range operators `..`, `..<`, `>..`, `>..<` (D13); legacy `.!`, `!.`, `!!`
  deprecated.
- [ ] `≡` reserved for **geometric congruence only**.
- [ ] List deprecated forms (`≠`, `==`, `!=`, `<>`) with E0010 notes.

### 1.6 `collections.html`  ← `spec/10-collections.md`
- [ ] Verify **Ordinal** type `(start){ id, id, … }` (commonly under-covered).
- [ ] Verify `$` end-anchor, 1-based indexing, and `$ - k` negative-relative.
- [ ] Verify set algebra `∩ ∪ \ Δ ⊂ ⊃ Σ` and `∈` membership.
- [ ] Verify matrix row-major + 2D indexing.
- [ ] Add diagnostic table `E1001`–`E1006`.

### 1.7 `processing.html`  ← `spec/11-processing.md`
- [ ] Verify boxing `[x]` and unboxing `Type(boxed)`.
- [ ] Verify quantifiers `∀`/`exists` and `∃`/`exists`.
- [ ] Verify pipelines `>>`, `.map/.filter/.reduce`, aggregates.
- [ ] Verify deconstruction `*tail` and matrix row/col slice `M[1,*]`, `M[*,2]`.
- [ ] Add diagnostic table `E1101`–`E1106`.

### 1.8 `concurrency.html`  ← `spec/12-concurrency.md`
- [ ] Verify `begin`/`wait`, thread-safe reduction `+>`, coroutines `yield`,
  channel extraction `yield var << task`.
- [ ] Verify worker exception isolation + `$trial` handle re-raise at `wait`.
- [ ] Add diagnostic table `E1201`–`E1205`.

### 1.9 `graphics.html`  ← `spec/13-graphics.md`
- [ ] Verify angular `G` type and `° ′ ″` literals.
- [ ] Verify `CRT/POL/VEC/CRC/SQR/PLG` primitives and `Canvas/Layer/Shape/Label`.
- [ ] Verify `draw/wipe/show/hide` and affine transforms.
- [ ] Verify `≡` congruence semantics live here (not as value equality).
- [ ] Add diagnostic table `E1301`–`E1305`.

### 1.10 `library.html`  ← `spec/14-library.md`
- [ ] Verify `$bee.sys` namespaces, `F` file handles, and tree-shaking linkage.
- [ ] Verify `entity.type()`, `size`, `length`, `capacity`.
- [ ] Verify `$error` code ranges (`1..199` system, `200+` user, `≤ -1` panic).
- [ ] Add diagnostic table `E1401`–`E1405`.

### 1.11 `structure.html`  ← `spec/04-structure.md`
- [ ] Verify `$` sigil system-variable table (`$bee_home`, `$pro_home`,
  `$max_precision`, …) is complete.
- [ ] Verify export prefix `.`, `with … do … done`, `alias`.
- [ ] Add diagnostic table (`E0401`–`E0405`).

### 1.12 `syntax.html`  ← `spec/01-lexical-structure.md`
- [ ] Verify UTF-8/rune handling description, BOM rejection (`E0101`).
- [ ] Verify identifier casing rules (lower variable vs upper type, single
  letter reserved types) — note the `spec/01` reserved-letter list (`A M L G`)
  conflicts with the `spec/05` catalogue (`A`=Alpha, `L`=Lambda, `G`=Angular);
  flag for reconciliation (see §3.5).
- [ ] Verify comment forms (`--`, `+- -+`, `(: :)`).
- [ ] Verify string literals (`'…'`, `"…"` interpolated `#(...)`, backtick raw)
  and markup blocks (`<text> <sql> <html> …`).
- [ ] Add lexical diagnostic table `E0101`–`E0105`.

---

## Priority 2 — New pages to make it a *complete* design document

These are additive, high-value pages for the "hundreds of students building
compilers" audience. Treat as proposals; confirm scope before authoring.

### 2.1 `diagnostics.html` (NEW)
- Consolidated, cross-referenced table of **every** diagnostic code
  (`E01xx`→`E14xx`, `W03xx`) grouped by module, with condition + resolution.
- Value: a compiler-builder's single source of truth for error surfaces.
- Requires resolving the E04xx collision first (§0.1 note, §3.5).

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
- The index table (lines ~16–34) marks D10 as "✅ Ratified" but the body and
  `MANIFEST.md` mark it "🟡 pending"; D11 status likewise drifts. It also omits
  D15. Reconcile the index with the body before any signature/automation
  depends on it.

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
- Remaining: audit structure/collections/processing/concurrency/graphics/library
  pages and inject their diagnostic tables now that registration is in place.

### 3.4 `features.html` overloaded + link drift
- `features.html` currently carries memory-model notes that belong in the new
  `memory.html`; trim and cross-link.
- [x] ~~Audit/remove the "Read next"/"previous" footer links.~~ Done: all
  `Read next/previous/more` footer blocks removed from `tutorial/*.html`.

### 3.5 Confirm source-of-truth framing
- The project rule (`GEMINI.md` §6) says `spec/` is the single source of truth
  and `tutorial/` is derived. The user now frames `tutorial/` as the design
  source-of-truth for students. Reconcile this framing explicitly so future
  agents know which direction facts flow (spec→tutorial) while the tutorial is
  still *complete enough* to stand alone for teaching.

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
