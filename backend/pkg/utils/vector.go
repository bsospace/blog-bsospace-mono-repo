package utils

import "math"

// CosineSimilarity คำนวณความคล้ายคลึงเชิงมุมระหว่างเวกเตอร์ 2 ตัว
// return ค่าอยู่ในช่วง [-1.0, 1.0] โดยที่ 1.0 คือเหมือนกันมากที่สุด
func CosineSimilarity(vec1, vec2 []float32) float64 {
	if len(vec1) != len(vec2) {
		return 0 // หรือจะ panic/log ก็ได้
	}

	var dotProduct, normA, normB float64

	for i := 0; i < len(vec1); i++ {
		a := float64(vec1[i])
		b := float64(vec2[i])

		dotProduct += a * b
		normA += a * a
		normB += b * b
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}
