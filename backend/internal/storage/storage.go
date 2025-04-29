package storage

import (
	"sync"
)

type Chunk struct {
	Text      string
	Embedding []float32
}

var (
	mu     sync.RWMutex
	Chunks []Chunk
)

func SaveChunksWithEmbeddings(chunks []Chunk) {
	mu.Lock()
	defer mu.Unlock()
	Chunks = chunks
}

func GetChunks() []Chunk {
	mu.RLock()
	defer mu.RUnlock()
	return Chunks
}
