package markdown

import (
	"bytes"
	"html/template"
	"strings"

	htmlchroma "github.com/alecthomas/chroma/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

var md = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,
		extension.DefinitionList,
		extension.Footnote,
		highlighting.NewHighlighting(
			highlighting.WithStyle("monokai"),
			highlighting.WithFormatOptions(
				htmlchroma.WithLineNumbers(true),
			),
		),
		&EmojiExtention{},
	),
	goldmark.WithParserOptions(
		parser.WithAutoHeadingID(),
		parser.WithAttribute(),
	),
	goldmark.WithRendererOptions(
		html.WithUnsafe(),
	),
)

func Convert(content []byte) []byte {
	buf := &bytes.Buffer{}
	err := md.Convert([]byte(strings.TrimSpace(string(content))), buf)
	if err != nil {
		buf.Reset()
		buf.Write([]byte("Error when creating markdown: " + err.Error()))
	}
	return buf.Bytes()
}

func ConvertSafeTemplate(content string) template.HTML {
	return template.HTML(Convert([]byte(content)))
}
