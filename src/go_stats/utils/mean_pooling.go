package utils

func MeanPooling(data []float32, mask []int64, seqLen int, hiddenSize int) []float32 {
	sentenceVector := make([]float32, hiddenSize)
	var count float32 = 0.0
	for i := 0; i < seqLen; i++ {
		if mask[i] == 0 {
			continue
		}
		start := i * hiddenSize
		end := start + hiddenSize
		tokenVector := data[start:end]
		for j := 0; j < hiddenSize; j++ {
			sentenceVector[j] += tokenVector[j]
		}
		count++
	}
	if count > 0 {
		for j := 0; j < hiddenSize; j++ {
			sentenceVector[j] /= count
		}
	}
	return sentenceVector
}
