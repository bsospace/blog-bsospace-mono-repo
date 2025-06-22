package tiptap

import (
	"encoding/json"
	"strings"
)

// ExtractTextFromTiptap converts TipTap JSON to formatted Markdown string.
// It preserves structure such as headings, paragraphs, and code blocks.
func ExtractTextFromTiptap(content string) string {
	var doc map[string]interface{}
	if err := json.Unmarshal([]byte(content), &doc); err != nil {
		return content // fallback to raw if invalid JSON
	}
	var builder strings.Builder
	walkTiptap(doc, &builder, "")
	return builder.String()
}

func walkTiptap(node interface{}, builder *strings.Builder, parentType string) {
	switch node := node.(type) {
	case map[string]interface{}:
		nodeType := node["type"]
		content := node["content"]
		var nodeTypeStr string

		// Handle specific node types
		if nodeTypeStr, ok := nodeType.(string); ok {
			switch nodeTypeStr {
			case "heading":
				level := 1
				if attrs, ok := node["attrs"].(map[string]interface{}); ok {
					if l, ok := attrs["level"].(float64); ok {
						level = int(l)
					}
				}
				builder.WriteString(strings.Repeat("#", level) + " ")
			case "paragraph":
				// Add line break before and after paragraph
				builder.WriteString("\n")
			case "hardBreak":
				builder.WriteString("\n")
			case "codeBlock":
				lang := ""
				if attrs, ok := node["attrs"].(map[string]interface{}); ok {
					if l, ok := attrs["language"].(string); ok {
						lang = l
					}
				}
				builder.WriteString("\n```" + lang + "\n")
				// Special handling: codeBlock content is just text
				if contentArr, ok := content.([]interface{}); ok {
					for _, child := range contentArr {
						if childMap, ok := child.(map[string]interface{}); ok {
							if text, ok := childMap["text"].(string); ok {
								builder.WriteString(text + "\n")
							}
						}
					}
				}
				builder.WriteString("```\n")
				return // codeBlock handled fully, no need to recurse
			}
		}

		// Handle text
		if text, ok := node["text"].(string); ok {
			builder.WriteString(text)
		}

		// Walk children
		if contentArr, ok := content.([]interface{}); ok {
			for _, child := range contentArr {
				walkTiptap(child, builder, "")
			}
		}

		// Extra spacing after certain block types
		if nodeTypeStr == "paragraph" || nodeTypeStr == "heading" {
			builder.WriteString("\n")
		}

	case []interface{}:
		for _, child := range node {
			walkTiptap(child, builder, parentType)
		}
	}
}
