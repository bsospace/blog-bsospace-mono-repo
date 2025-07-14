package tiptap

import (
	"math"
	"strings"

	"golang.org/x/net/html"
)

// Strip all HTML tags and count words
func EstimateReadTimeFromHTML(htmlContent string) int {
	text := extractTextFromHTML(htmlContent)
	words := strings.Fields(text)
	wordCount := len(words)

	// Average reading speed: 200 words per minute
	minutes := math.Ceil(float64(wordCount) / 200.0)
	return int(minutes)
}

func extractTextFromHTML(htmlStr string) string {
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		return htmlStr // fallback: return raw if can't parse
	}
	var sb strings.Builder
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			sb.WriteString(n.Data + " ")
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return sb.String()
}
