package tiptap

import (
	"encoding/json"
	"strings"
)

// Recursively extract all "text" fields from TipTap JSON
func ExtractTextFromTiptap(content string) string {
	var doc map[string]interface{}
	if err := json.Unmarshal([]byte(content), &doc); err != nil {
		return content // fallback to raw if error
	}
	var builder strings.Builder
	walkTiptap(doc, &builder)
	return builder.String()
}

func walkTiptap(node interface{}, builder *strings.Builder) {
	switch node := node.(type) {
	case map[string]interface{}:
		// Check for "text"
		if text, ok := node["text"].(string); ok {
			builder.WriteString(text + " ")
		}
		// Walk children
		if content, ok := node["content"].([]interface{}); ok {
			for _, child := range content {
				walkTiptap(child, builder)
			}
		}
	case []interface{}:
		for _, child := range node {
			walkTiptap(child, builder)
		}
	}
}
