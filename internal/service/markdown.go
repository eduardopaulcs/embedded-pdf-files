package service

import (
	"bytes"
	"html/template"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
)

func MarkdownToHTML(input []byte) (template.HTML, error) {
	var buf bytes.Buffer
	md := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)
	if err := md.Convert(input, &buf); err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil
}
