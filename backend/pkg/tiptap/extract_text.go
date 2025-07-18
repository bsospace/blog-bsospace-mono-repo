package tiptap

import (
	"encoding/json"
	"fmt"
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

// ExtractTextFromTiptap converts TipTap JSON content into Markdown
func ExtractTextFromTiptapToMD(content string) string {
	var doc map[string]interface{}
	if err := json.Unmarshal([]byte(content), &doc); err != nil {
		return content // fallback to raw if invalid
	}
	var builder strings.Builder
	walkTiptapToMD(doc, &builder, "")
	return strings.TrimSpace(builder.String())
}

func walkTiptapToMD(node interface{}, builder *strings.Builder, parentType string) {
	switch n := node.(type) {
	case map[string]interface{}:
		nodeType, _ := n["type"].(string)
		attrs, _ := n["attrs"].(map[string]interface{})
		content, _ := n["content"]

		// Handle block elements
		switch nodeType {
		case "heading":
			level := 1
			if l, ok := attrs["level"].(float64); ok {
				level = int(l)
			}
			builder.WriteString("\n" + strings.Repeat("#", level) + " ")
		case "paragraph":
			builder.WriteString("\n")
		case "bulletList":
			parentType = "ul"
		case "orderedList":
			parentType = "ol"
		case "listItem":
			prefix := "- "
			if parentType == "ol" {
				prefix = "1. "
			}
			builder.WriteString("\n" + prefix)
		case "codeBlock":
			lang := ""
			if l, ok := attrs["language"].(string); ok {
				lang = l
			}
			builder.WriteString("\n```" + lang + "\n")
			if arr, ok := content.([]interface{}); ok {
				for _, c := range arr {
					if cm, ok := c.(map[string]interface{}); ok {
						if text, ok := cm["text"].(string); ok {
							builder.WriteString(text + "\n")
						}
					}
				}
			}
			builder.WriteString("```\n")
			return
		case "image":
			src, _ := attrs["src"].(string)
			alt, _ := attrs["alt"].(string)
			builder.WriteString(fmt.Sprintf("\n![%s](%s)\n", alt, src))
			return
		case "hardBreak":
			builder.WriteString("  \n")
			return
		}

		// Handle text with formatting (marks)
		if text, ok := n["text"].(string); ok {
			text = applyMarks(text, n["marks"])
			builder.WriteString(text)
			return
		}

		// Recurse into children
		if arr, ok := content.([]interface{}); ok {
			for _, c := range arr {
				walkTiptap(c, builder, parentType)
			}
		}

		// Add spacing for some block types
		if nodeType == "paragraph" || strings.HasPrefix(nodeType, "heading") {
			builder.WriteString("\n")
		}

	case []interface{}:
		for _, c := range n {
			walkTiptap(c, builder, parentType)
		}
	}
}

func applyMarks(text string, marksInterface interface{}) string {
	if marksInterface == nil {
		return text
	}
	marks, ok := marksInterface.([]interface{})
	if !ok {
		return text
	}

	for _, mark := range marks {
		if m, ok := mark.(map[string]interface{}); ok {
			switch m["type"] {
			case "bold":
				text = "**" + text + "**"
			case "italic":
				text = "_" + text + "_"
			case "code":
				text = "`" + text + "`"
			case "underline":
				text = "<u>" + text + "</u>"
			case "strike":
				text = "~~" + text + "~~"
			}
		}
	}
	return text
}
