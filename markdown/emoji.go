package markdown

import (
	"fmt"
	"regexp"

	"github.com/kyokomi/emoji"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type Emoji struct {
	ast.BaseInline
	Value []byte
}

func (n *Emoji) Dump(source []byte, level int) {
	m := map[string]string{
		"Value": fmt.Sprintf("%v", n.Value),
	}
	ast.DumpHelper(n, source, level, m, nil)
}

var KindEmoji = ast.NewNodeKind("Emoji")

func (n *Emoji) Kind() ast.NodeKind {
	return KindEmoji
}

func NewEmoji(value []byte) *Emoji {
	return &Emoji{
		Value: value,
	}
}

var emojiRegexp = regexp.MustCompile(`^:[A-Za-z0-9_-]*:`)

type emojiParser struct {
}

var defaultReferParser = &emojiParser{}

func NewEmojiParser() parser.InlineParser {
	return defaultReferParser
}

func (s *emojiParser) Trigger() []byte {
	return []byte{':'}
}

func (s *emojiParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, segment := block.PeekLine()
	m := emojiRegexp.FindSubmatchIndex(line)
	if m == nil {
		return nil
	}

	block.Advance(m[1])
	node := NewEmoji(block.Value(text.NewSegment(segment.Start+1, segment.Start+m[1]-1)))
	return node
}

func (s *emojiParser) CloseBlock(parent ast.Node, pc parser.Context) {
	// nothing to do
}

type EmojiRenderer struct {
}

func NewEmojiRenderer() renderer.NodeRenderer {
	r := &EmojiRenderer{
	}

	return r
}

// RegisterFuncs implements NodeRenderer.RegisterFuncs.
func (r *EmojiRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindEmoji, r.renderEmoji)
}

func (r *EmojiRenderer) renderEmoji(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*Emoji)
	key := ":" + string(n.Value) + ":"
	if r.notAllowedBlock(n) {
		w.WriteString(key)
		return ast.WalkContinue, nil
	}

	result, ok := emoji.CodeMap()[key]
	if !ok {
		w.WriteString(key)
		return ast.WalkContinue, nil
	}
	w.WriteString(result)
	return ast.WalkContinue, nil
}

func (r *EmojiRenderer) notAllowedBlock(node ast.Node) bool {
	if node == nil || node.Parent() == nil {
		return false
	}
	kind := node.Parent().Kind()
	if kind == ast.KindCodeBlock ||
		kind == ast.KindBlockquote {
		return true
	}
	return r.notAllowedBlock(node.Parent())
}

type EmojiExtention struct {
}

func (EmojiExtention) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(NewEmojiParser(), 500),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewEmojiRenderer(), 500),
	))
}
