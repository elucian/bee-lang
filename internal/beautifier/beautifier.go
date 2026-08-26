// internal/beautifier/beautifier.go
// Purpose: Native source code beautifier for the Bee Programming Language.
// Responsibility: Enforces strict 2-space block indentation, formats block headers,
//                 aligns trailing comments, and auto-corrects simple syntax errors.
// Core Architectural Strategy: Uses token stream analysis and line-level parsing
//                              to rewrite Bee source code deterministically with zero third-party dependencies.

package beautifier

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Option defines a functional configuration option for the Beautifier.
type Option func(*Beautifier)

// Beautifier holds configuration settings for source code formatting.
type Beautifier struct {
	indentSpaces  int
	alignComments bool
	autoFix       bool
}

// New creates a new Beautifier instance with default 2-space indentation.
// Inputs: opts (...Option) - optional configuration functions.
// Outputs: *Beautifier - initialized beautifier engine.
func New(opts ...Option) *Beautifier {
	b := &Beautifier{
		indentSpaces:  2,
		alignComments: true,
		autoFix:       true,
	}
	for _, opt := range opts {
		opt(b)
	}
	return b
}

// FormatSource formats raw Bee source code and returns the beautified string.
// Inputs: input (string) - raw Bee source code.
// Outputs: (string, error) - beautified source code or formatting error.
func (b *Beautifier) FormatSource(input string) (string, error) {
	cleanInput := strings.ReplaceAll(input, "\r\n", "\n")
	lines := strings.Split(cleanInput, "\n")

	var formattedLines []string
	indentLevel := 0

	for lineNo, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if len(formattedLines) > 0 && formattedLines[len(formattedLines)-1] != "" {
				formattedLines = append(formattedLines, "")
			}
			continue
		}

		trimmed = b.autoFixLine(trimmed)

		isCloser := b.isBlockCloser(trimmed)
		isMiddle := b.isBlockMiddle(trimmed)

		if isCloser {
			indentLevel--
			if indentLevel < 0 {
				indentLevel = 0
				fmt.Fprintf(os.Stderr, "DEBUG: [Beautifier] Unmatched block terminator at line %d: %s\n", lineNo+1, trimmed)
			}
		}

		// Indentation calculation:
		// Regular openers (rule, if, trial, match) increment indentLevel for their body.
		// Middle/branch clauses (else, try, case, miss, final, when, other) sit at the parent block's inner indentation level (indentLevel - 1).
		visualIndent := indentLevel
		if isMiddle {
			visualIndent = indentLevel - 1
			if visualIndent < 0 {
				visualIndent = 0
			}
		}

		pad := strings.Repeat(" ", visualIndent*b.indentSpaces)
		formattedLines = append(formattedLines, pad+trimmed)

		isOpener := b.isBlockOpener(trimmed)
		if isOpener {
			indentLevel++
		}
	}

	if b.alignComments {
		formattedLines = b.alignCommentsInBlock(formattedLines)
	}

	result := strings.Join(formattedLines, "\n") + "\n"
	return result, nil
}

// autoFixLine applies deterministic auto-corrections such as implicit multiplication.
// Inputs: line (string) - single line of code.
// Outputs: string - auto-corrected line.
func (b *Beautifier) autoFixLine(line string) string {
	if !b.autoFix {
		return line
	}

	reNumParen := regexp.MustCompile(`(\b\d+)\s*\(([^)]+)\)`)
	line = reNumParen.ReplaceAllString(line, `$1 * ($2)`)

	return line
}

// alignCommentsInBlock aligns trailing comments across formatted lines to a consistent column.
// Inputs: lines ([]string) - list of formatted lines.
// Outputs: []string - list of lines with aligned trailing comments.
func (b *Beautifier) alignCommentsInBlock(lines []string) []string {
	var aligned []string
	minCommentCol := 40

	for _, line := range lines {
		idx := strings.Index(line, "--")
		if idx > 0 {
			codePart := strings.TrimRight(line[:idx], " \t")
			if len(codePart)+2 > minCommentCol {
				minCommentCol = len(codePart) + 2
			}
		}
	}

	for _, line := range lines {
		idx := strings.Index(line, "--")
		if idx > 0 {
			codePart := strings.TrimRight(line[:idx], " \t")
			commentPart := line[idx:]
			padding := strings.Repeat(" ", max(1, minCommentCol-len(codePart)))
			aligned = append(aligned, codePart+padding+commentPart)
		} else {
			aligned = append(aligned, line)
		}
	}
	return aligned
}

// isBlockOpener checks if a line opens a new block.
// Inputs: line (string) - trimmed line of code.
// Outputs: bool - true if the line opens a block.
func (b *Beautifier) isBlockOpener(line string) bool {
	if b.endsWithColon(line) || strings.HasSuffix(line, "do") || strings.HasSuffix(line, "do:") {
		if b.isBlockMiddle(line) {
			return false
		}
		return true
	}
	openerKeywords := []string{"rule ", "if ", "while ", "for ", "cycle", "match", "trial", "with "}
	for _, kw := range openerKeywords {
		if strings.HasPrefix(line, kw) && !strings.HasSuffix(line, ";") {
			return true
		}
	}
	return false
}

// isBlockMiddle checks if a line is a block continuation keyword (like else, case, miss, final, try, when, other).
// Inputs: line (string) - trimmed line of code.
// Outputs: bool - true if the line is a block middle keyword.
func (b *Beautifier) isBlockMiddle(line string) bool {
	return strings.HasPrefix(line, "else") || strings.HasPrefix(line, "case") || strings.HasPrefix(line, "miss") || strings.HasPrefix(line, "final") || strings.HasPrefix(line, "try") || strings.HasPrefix(line, "when") || strings.HasPrefix(line, "other")
}

// isBlockCloser checks if a line closes a block.
// Inputs: line (string) - trimmed line of code.
// Outputs: bool - true if the line closes a block.
func (b *Beautifier) isBlockCloser(line string) bool {
	return strings.HasPrefix(line, "done") || strings.HasPrefix(line, "repeat") || strings.HasPrefix(line, "return")
}

func (b *Beautifier) endsWithColon(s string) bool {
	trimmed := strings.TrimRight(s, " \t")
	return strings.HasSuffix(trimmed, ":")
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
