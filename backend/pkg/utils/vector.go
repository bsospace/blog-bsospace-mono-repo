package utils

import "math"

// CosineSimilarity คำนวณความใกล้เคียงของ 2 vectors
func CosineSimilarity(vec1, vec2 []float32) float64 {
	var dotProduct float64
	var normA float64
	var normB float64

	for i := 0; i < len(vec1) && i < len(vec2); i++ {
		dotProduct += float64(vec1[i] * vec2[i])
		normA += float64(vec1[i] * vec1[i])
		normB += float64(vec2[i] * vec2[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}
