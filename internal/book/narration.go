package book

import (
	"fmt"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// Chapters derives speech only. Source, prepared passages and the editable master
// keep their Markdown. Parse the complete master before previewing or chunking so
// delimiters and reference links can span passage/chunk boundaries.
func Chapters(master string) []Chapter {
	source := []byte(master)
	md := goldmark.New(goldmark.WithExtensions(extension.Table, extension.Strikethrough, extension.TaskList))
	root := md.Parser().Parse(text.NewReader(source))
	var chapters []Chapter
	title := "Listening Edition"
	var body strings.Builder
	ordered := map[*ast.List]int{}
	flush := func() {
		if speech := strings.TrimSpace(body.String()); speech != "" {
			chapters = append(chapters, Chapter{Title: title, Text: speech})
		}
		body.Reset()
	}
	_ = ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			if h, ok := n.(*ast.Heading); ok && h.Level == 1 && n.Parent() == root {
				flush()
				title = inlineSpeech(n, source)
				if title != "" {
					body.WriteString(headingPause(title) + "\n\n")
				}
				return ast.WalkSkipChildren, nil
			}
			if speechLeaf(&body, n, source) {
				return ast.WalkSkipChildren, nil
			}
			switch v := n.(type) {
			case *ast.CodeBlock, *ast.FencedCodeBlock:
				body.Write(n.Lines().Value(source)) // Code content is literal; only its fence is silent.
				body.WriteString("\n\n")
				return ast.WalkSkipChildren, nil
			case *ast.HTMLBlock:
				// Unknown/raw HTML remains literal text, never rendered or executed.
				body.Write(v.Lines().Value(source))
				if v.HasClosure() {
					body.Write(v.ClosureLine.Value(source))
				}
			case *ast.List:
				ordered[v] = v.Start
			case *ast.ListItem:
				if list, ok := n.Parent().(*ast.List); ok && list.IsOrdered() {
					fmt.Fprintf(&body, "%d. ", ordered[list])
					ordered[list]++
				}
			case *extast.TaskCheckBox:
				if v.IsChecked {
					body.WriteString("Completed: ")
				} else {
					body.WriteString("Not completed: ")
				}
			}
		} else {
			switch n.(type) {
			case *ast.Paragraph, *ast.Heading, *ast.ListItem, *ast.Blockquote, *ast.ThematicBreak, *ast.HTMLBlock:
				body.WriteString("\n\n")
			case *ast.TextBlock, *extast.TableHeader, *extast.TableRow:
				body.WriteByte('\n')
			case *extast.TableCell:
				if n.NextSibling() != nil {
					body.WriteString("; ")
				}
			}
		}
		return ast.WalkContinue, nil
	})
	flush()
	return chapters
}

func headingPause(s string) string {
	if strings.ContainsAny(s[len(s)-1:], ".!?:;") {
		return s
	}
	return s + "."
}

func inlineSpeech(n ast.Node, source []byte) string {
	var out strings.Builder
	_ = ast.Walk(n, func(child ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && speechLeaf(&out, child, source) {
			return ast.WalkSkipChildren, nil
		}
		return ast.WalkContinue, nil
	})
	return strings.TrimSpace(out.String())
}

// Write visible inline text, not Markdown delimiters or a link's destination.
// Real symbols, escaped punctuation, citations, bare URLs and code remain text.
func speechLeaf(out *strings.Builder, n ast.Node, source []byte) bool {
	decode := func(value []byte, raw bool) {
		if !raw {
			value = util.UnescapePunctuations(value)
			value = util.ResolveNumericReferences(value)
			value = util.ResolveEntityNames(value)
		}
		out.Write(value)
	}
	switch v := n.(type) {
	case *ast.Text:
		decode(v.Value(source), v.IsRaw())
		if v.SoftLineBreak() || v.HardLineBreak() {
			out.WriteByte('\n')
		}
	case *ast.String:
		decode(v.Value, v.IsRaw())
	case *ast.CodeSpan:
		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
			if t, ok := child.(*ast.Text); ok {
				out.WriteString(strings.ReplaceAll(string(t.Value(source)), "\n", " "))
			}
		}
	case *ast.AutoLink:
		decode(v.Label(source), false)
	case *ast.RawHTML:
		out.Write(v.Segments.Value(source))
	default:
		return false
	}
	return true
}
