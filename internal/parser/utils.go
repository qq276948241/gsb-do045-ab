package parser

import (
	"fmt"
	"strings"
	"unicode"

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
	// Slug generation is a direct rendering of the shared canonical heading
	// form: words are joined with single hyphens. Every other stage (heading
	// parsing, link checking, heading rules) compares against this same form.
	return strings.ReplaceAll(NormalizeHeading(text), " ", "-")
}

func isInternalLink(url string) bool {
	return !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://")
}

// NormalizeHeading reduces a heading (or any text compared as a heading) to a
// single canonical form shared by heading parsing, anchor generation, internal
// link checking, and heading rules:
//  1. Convert to lowercase (case never matters)
//  2. Treat spaces and hyphens as word separators
//  3. Collapse runs of separators (including multiple consecutive spaces)
//  4. Keep every other character, including commas
//
// The result uses single spaces between words; GenerateSlug renders the same
// form with hyphens. An empty result means the text carries no anchor (e.g.
// punctuation or whitespace only).
func NormalizeHeading(text string) string {
	var words []string
	var current strings.Builder
	flush := func() {
		if current.Len() > 0 {
			words = append(words, current.String())
			current.Reset()
		}
	}

	for _, r := range strings.ToLower(text) {
		if unicode.IsSpace(r) || r == '-' {
			flush()
			continue
		}
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '_', r == ',':
			current.WriteRune(r)
		default:
			// Other punctuation and unsupported characters do not take part
			// in the anchor; commas above are deliberately retained.
		}
	}
	flush()

	return strings.Join(words, " ")
}

// NormalizeAnchor applies the same canonicalization as NormalizeHeading to the
// fragment of an internal link. Link fragments conventionally use hyphens, so
// spaces and hyphens collapse to the same separators; case and repeated spaces
// are folded away and commas are preserved.
func NormalizeAnchor(anchor string) string {
	return NormalizeHeading(anchor)
}
