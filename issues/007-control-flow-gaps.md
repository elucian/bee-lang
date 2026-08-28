# Issue: Missing Control Flow Syntax in Specification
The formal specification `spec/02-statements.md` is missing the definitions for `start` (local scope creation) and `pass` (execution transfer) control flow statements, both of which are documented in the web pages.

## Impact
- Incomplete EBNF grammar prevents correct parsing of `start` and `pass` statements.
- Developer reference `spec/` is inconsistent with language capabilities described in `web/`.

## Requirements
- Add `start` block definition to control flow section.
- Add `pass` keyword to transfer statements EBNF.
