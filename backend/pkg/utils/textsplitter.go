package utils

import (
	"strings"
)

// ChunkText splits large text into chunks with approximately n words per chunk
func ChunkText(text string, chunkSize int) []string {
	words := strings.Fields(text)
	var chunks []string
	for i := 0; i < len(words); i += chunkSize {
		end := i + chunkSize
		if end > len(words) {
			end = len(words)
		}
		chunks = append(chunks, strings.Join(words[i:end], " "))
	}
	return chunks
}
