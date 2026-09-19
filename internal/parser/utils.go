package parser

import (
	"fmt"
	"strings"

	"github.com/yuin/goldmark/ast"
	east "github.com/yuin/goldmark/extension/ast"
)

func getPosition(node ast.Node, content []byte) (line, col int) {
	// Try to get position from different node types
	switch n := node.(type) {
	case *ast.Heading:
		if n.Lines().Len() > 0 {
			return calculateLineColumn(content, n.Lines().At(0).Start)
		}
	case *ast.FencedCodeBlock:
		if n.Lines().Len() > 0 {
			return calculateLineColumn(content, n.Lines().At(0).Start)
		}
	case *ast.Text:
		return calculateLineColumn(content, n.Segment.Start)
	case *ast.List:
		if n.Lines().Len() > 0 {
			return calculateLineColumn(content, n.Lines().At(0).Start)
		}
	case *east.Table:
		if n.Lines().Len() > 0 {
			return calculateLineColumn(content, n.Lines().At(0).Start)
		}
	case *ast.Paragraph:
		if n.Lines().Len() > 0 {
			return calculateLineColumn(content, n.Lines().At(0).Start)
		}
	}

	// Fallback to finding the first text node
	var firstOffset int
	if err := ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if t, ok := n.(*ast.Text); ok && firstOffset == 0 {
			firstOffset = t.Segment.Start
			return ast.WalkStop, nil
		}
		return ast.WalkContinue, nil
	}); err != nil {
		fmt.Printf("Error getting position: %v\n", err)
		return 1, 1 // Default to line 1, column 1 on error
	}

	if firstOffset > 0 {
		return calculateLineColumn(content, firstOffset)
	}

	// Default fallback
	return 1, 1
}

func calculateLineColumn(content []byte, offset int) (line, col int) {
	line = 1
	col = 1

	for i := 0; i < offset && i < len(content); i++ {
		if content[i] == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}

	return line, col
}

func GenerateSlug(text string) string {
	return NormalizeAnchor(text)
}

// NormalizeText is the single normalization used everywhere headings or
// internal anchors are compared:
//  1. Trim surrounding whitespace
//  2. Fold consecutive spaces into a single space
//  3. Convert to lowercase
//
// Commas (and other punctuation) are preserved; only spaces and case are
// folded. Heading parsing, anchor generation/slug indexing, internal link
// checks, and heading uniqueness rules must all compare via this function so
// they never disagree.
func NormalizeText(text string) string {
	return strings.ToLower(strings.Join(strings.Fields(text), " "))
}

// NormalizeAnchor reduces a heading or an internal link target to the anchor
// form used for comparison. It applies NormalizeText first, then maps the
// normalized spaces to hyphens. Commas are retained, so "Hello, World" and a
// link to "#hello,-world" resolve to the same anchor "hello,-world".
//
// An input containing only whitespace (e.g. the target after a bare "# ")
// normalizes to the empty string, which callers must treat as an invalid
// anchor rather than a match.
func NormalizeAnchor(text string) string {
	normalized := NormalizeText(text)
	if normalized == "" {
		return ""
	}
	return strings.ReplaceAll(normalized, " ", "-")
}

func isInternalLink(url string) bool {
	return !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://")
}
