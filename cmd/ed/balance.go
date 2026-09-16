package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

const balanceUsage = `bee-ed balance - validate HTML/Markdown tag balance

USAGE:
  bee-ed balance <file>

Recognizes HTML tags, HTML5 void elements (br, img, hr, ... built-in), and
&nbsp;-style self-closing "<tag/>" in Markdown. Template braces and code
fences are tolerated. Comments (<!-- -->), CDATA, and Doctype are skipped.
Prints a summary and exits non-zero if any tag is unbalanced.

FLAGS:
  --help, -h  Show this help.
`

// htmlVoidTags are HTML5 elements that never need a closing tag.
var htmlVoidTags = map[string]bool{
	"area": true, "base": true, "br": true, "col": true, "embed": true,
	"hr": true, "img": true, "input": true, "link": true, "meta": true,
	"param": true, "source": true, "track": true, "wbr": true,
}

var (
	reComment = regexp.MustCompile(`(?s)<!--.*?-->`)
	reDoctype = regexp.MustCompile(`(?i)<!DOCTYPE[^>]*>`)
	reCDATA   = regexp.MustCompile(`(?s)<!\[CDATA\[.*?\]\]>`)
	// Any <tagname ...> or </tagname> ; handles attributes with quoted '=' .
	reTag = regexp.MustCompile(`</?[a-zA-Z][^>]*>`)
	// rawTextTags are HTML raw-text / RCDATA content elements whose inner text
	// is NOT parsed as markup. Their contents may quote arbitrary tags (e.g. a
	// <pre> sample documenting <html>), which must not count toward balance.
	rawTextTags = map[string]bool{
		"pre": true, "code": true, "script": true, "style": true, "textarea": true,
	}
)

// maskRawText blanks the inner content of raw-text HTML elements so tags shown
// as documentation data (e.g. HTML samples inside <pre>/<code>) are not counted
// as markup. Go's regexp (RE2) has no backreferences, so this is a hand-rolled
// scanner: on an opening raw-text tag it skips to the matching case-insensitive
// close tag and blanks the interior, preserving byte lengths for stable indices.
func maskRawText(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	i := 0
	n := len(s)
	for i < n {
		open := reTag.FindStringIndex(s[i:])
		if open == nil {
			b.WriteString(s[i:])
			break
		}
		// Copy text before this tag verbatim.
		start := i + open[0]
		end := i + open[1]
		b.WriteString(s[i:start])
		tok := s[start:end]
		name, closing, _ := parseTag(tok)
		if !closing && rawTextTags[strings.ToLower(name)] {
			b.WriteString(tok) // preserve the opening tag
			// Find the matching closing tag, honoring nesting of the same element.
			// Only the INTERIOR content is blanked, so the structural open/close
			// tags still count toward balance while any fake tags inside (e.g. an
			// HTML sample documenting <html>) are excluded from inspection.
			depth := 1
			cur := end
			for depth > 0 {
				idx := reTag.FindStringIndex(s[cur:])
				if idx == nil {
					// No close tag: the raw-text element runs to EOF; blank the rest.
					b.WriteString(strings.Repeat("x", n-cur))
					cur = n
					break
				}
				at := cur + idx[0]
				to := cur + idx[1]
				innerTok := s[at:to]
				iname, iclosing, _ := parseTag(innerTok)
				if strings.EqualFold(iname, name) {
					if iclosing {
						depth--
					} else {
						depth++
					}
				}
				if depth == 0 {
					// Preserve the close tag; blank the interior content before it.
					b.WriteString(strings.Repeat("x", at-cur))
					b.WriteString(innerTok)
					cur = to
					break
				}
				cur = to
			}
			i = cur
			continue
		}
		b.WriteString(tok)
		i = end
	}
	return b.String()
}

func runBalance(args []string) error {
	var file string
	for _, a := range args {
		switch a {
		case "--help", "-h":
			fmt.Fprint(os.Stdout, balanceUsage)
			return nil
		default:
			if strings.HasPrefix(a, "-") {
				return fmt.Errorf("unknown flag %q\n\n%s", a, balanceUsage)
			}
			if file == "" {
				file = a
			} else {
				return fmt.Errorf("too many arguments (got %q)", a)
			}
		}
	}
	if file == "" {
		return fmt.Errorf("missing <file>\n\n" + balanceUsage)
	}
	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	problems, open := checkBalance(string(content))
	if len(problems) == 0 && len(open) == 0 {
		fmt.Fprintf(os.Stdout, "balance %s: OK (%d tags matched)\n", file, countTags(string(content)))
		return nil
	}
	for _, p := range problems {
		fmt.Fprintf(os.Stderr, "  ! %s\n", p)
	}
	for _, o := range open {
		fmt.Fprintf(os.Stderr, "  ! unclosed <%s>\n", o)
	}
	return fmt.Errorf("%d balance problem(s) in %q", len(problems)+len(open), file)
}

// checkBalance strips comments/doctype/CDATA then walks tags on a stack.
// It returns a list of mismatch problems and a list of unclosed tag names.
func checkBalance(content string) ([]string, []string) {
	body := reComment.ReplaceAllString(content, "")
	body = reDoctype.ReplaceAllString(body, "")
	body = reCDATA.ReplaceAllString(body, "")
	// Ignore tags inside raw-text/HTML-content elements (documentation data).
	body = maskRawText(body)
	// Detect code fences + markdown inline code to skip tags inside them.
	body = maskCodeFences(body)

	var stack []string
	var problems []string
	for _, m := range reTag.FindAllStringIndex(body, -1) {
		tok := body[m[0]:m[1]]
		name, closing, selfClose := parseTag(tok)
		if name == "" {
			continue
		}
		// HTML5 void element: ignore close expectations.
		if htmlVoidTags[strings.ToLower(name)] {
			continue
		}
		if selfClose {
			continue // <tag/> is self-closing
		}
		if closing {
			if n := len(stack); n > 0 && strings.EqualFold(stack[n-1], name) {
				stack = stack[:n-1]
			} else {
				problems = append(problems, fmt.Sprintf("mismatched </%s> (expected </%s> or none)", name, peek(stack)))
			}
		} else {
			stack = append(stack, name)
		}
	}
	return problems, stack
}

func peek(s []string) string {
	if len(s) == 0 {
		return ""
	}
	return s[len(s)-1]
}

// maskCodeFences replaces Markdown ``` blocks and `inline` spans with spaces,
// so a parser/language reference inside them does not trip tag balance.
func maskCodeFences(s string) string {
	var b strings.Builder
	// Handle fenced blocks ``` ... ```
	parts := strings.Split(s, "```")
	for i, p := range parts {
		if i%2 == 1 {
			b.WriteString(strings.Repeat("x", len(p)))
		} else {
			b.WriteString(p)
		}
	}
	out := b.String()
	// Inline `code` spans: replace with underscores, keeping their length.
	reInline := regexp.MustCompile("`[^`\n]*`")
	return reInline.ReplaceAllStringFunc(out, func(m string) string {
		return strings.Repeat("_", len(m))
	})
}

// parseTag decomposes a tag token into name, closing flag, and self-close flag.
func parseTag(tok string) (name string, closing, selfClose bool) {
	t := tok
	if strings.HasPrefix(t, "</") {
		closing = true
		t = t[2:]
	} else {
		t = t[1:]
	}
	// Strip the trailing '=' of attribute needs only what's before whitespace.
	name = strings.Fields(t)[0]
	// name may carry an attribute char like '>' stuck on; strip non-letters.
	end := 0
	for end < len(name) && isTagNameChar(name[end]) {
		end++
	}
	name = name[:end]
	if len(t) >= 2 && t[len(t)-2] == '/' {
		selfClose = true
	}
	return name, closing, selfClose
}

func isTagNameChar(c byte) bool {
	return c == '-' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func countTags(content string) int {
	body := reComment.ReplaceAllString(content, "")
	return len(reTag.FindAllStringIndex(body, -1))
}
