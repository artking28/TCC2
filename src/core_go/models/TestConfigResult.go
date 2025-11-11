package models

import (
	"fmt"
	"math"
)

type TestConfigResult struct {
	// Velocidade de cálculo dos vetores vase dos documentos em bytes por segundo
	DocCalcBytesPerSecondAvgTime float64

	// Usando uma frase de 10 palavras, a similaridade media dentre as listas geradas por uma busca via Bert e via um algoritmo
	AvgSpearmanSim10 float64
	// Usando uma frase de 10 palavras, a menor similaridade dentre as listas geradas por uma busca via Bert e via um algoritmo
	MinSpearmanSim10 float64
	// Usando uma frase de 10 palavras, a maior similaridade dentre as listas geradas por uma busca via Bert e via um algoritmo
	MaxSpearmanSim10 float64
	// Tempo medio em millis para calcular o vetor base de frases de 10 palavras
	AvgTime10 int64
	// Menor tempo em millis para calcular o vetor base de frases de 10 palavras
	MinTime10 int64
	// Maior tempo em millis para calcular o vetor base de frases de 10 palavras
	MaxTime10 int64

	AvgSpearmanSim20 float64
	MinSpearmanSim20 float64
	MaxSpearmanSim20 float64
	AvgTime20        int64
	MinTime20        int64
	MaxTime20        int64

	AvgSpearmanSim40 float64
	MinSpearmanSim40 float64
	MaxSpearmanSim40 float64
	AvgTime40        int64
	MinTime40        int64
	MaxTime40        int64
}

func NewTestConfigResult() TestConfigResult {
	return TestConfigResult{
		DocCalcBytesPerSecondAvgTime: 0,

		AvgSpearmanSim10: 0,
		MinSpearmanSim10: math.MaxFloat64,
		MaxSpearmanSim10: -math.MaxFloat64,
		AvgTime10:        0,
		MinTime10:        math.MaxInt64,
		MaxTime10:        -math.MaxInt64,

		AvgSpearmanSim20: 0,
		MinSpearmanSim20: math.MaxFloat64,
		MaxSpearmanSim20: -math.MaxFloat64,
		AvgTime20:        0,
		MinTime20:        math.MaxInt64,
		MaxTime20:        -math.MaxInt64,

		AvgSpearmanSim40: 0,
		MinSpearmanSim40: math.MaxFloat64,
		MaxSpearmanSim40: -math.MaxFloat64,
		AvgTime40:        0,
		MinTime40:        math.MaxInt64,
		MaxTime40:        -math.MaxInt64,
	}
}

func (this *TestConfigResult) Push10(spearman float64, elapsedMillis int64) {
	this.AvgSpearmanSim10 += spearman
	if spearman < this.MinSpearmanSim10 {
		this.MinSpearmanSim10 = spearman
	}
	if spearman > this.MaxSpearmanSim10 {
		this.MaxSpearmanSim10 = spearman
	}

	this.AvgTime10 += elapsedMillis
	if elapsedMillis < this.MinTime10 {
		this.MinTime10 = elapsedMillis
	}
	if elapsedMillis > this.MaxTime10 {
		this.MaxTime10 = elapsedMillis
	}
}

func (this *TestConfigResult) Push20(spearman float64, elapsedMillis int64) {
	this.AvgSpearmanSim20 += spearman
	if spearman < this.MinSpearmanSim20 {
		this.MinSpearmanSim20 = spearman
	}
	if spearman > this.MaxSpearmanSim20 {
		this.MaxSpearmanSim20 = spearman
	}

	this.AvgTime20 += elapsedMillis
	if elapsedMillis < this.MinTime20 {
		this.MinTime20 = elapsedMillis
	}
	if elapsedMillis > this.MaxTime20 {
		this.MaxTime20 = elapsedMillis
	}
}

func (this *TestConfigResult) Push40(spearman float64, elapsedMillis int64) {
	this.AvgSpearmanSim40 += spearman
	if spearman < this.MinSpearmanSim40 {
		this.MinSpearmanSim40 = spearman
	}
	if spearman > this.MaxSpearmanSim40 {
		this.MaxSpearmanSim40 = spearman
	}

	this.AvgTime40 += elapsedMillis
	if elapsedMillis < this.MinTime40 {
		this.MinTime40 = elapsedMillis
	}
	if elapsedMillis > this.MaxTime40 {
		this.MaxTime40 = elapsedMillis
	}
}

func (this *TestConfigResult) String() (ret string) {

	ret += fmt.Sprintf("\tSpearmanSim10: avg=%.4f, min=%.4f, max=%.4f, time.avg=%d, time.min=%d, time.max=%d\n",
		this.AvgSpearmanSim10, this.MinSpearmanSim10, this.MaxSpearmanSim10,
		this.AvgTime10, this.MinTime10, this.MaxTime10)

	ret += fmt.Sprintf("\tSpearmanSim20: avg=%.4f, min=%.4f, max=%.4f, time.avg=%d, time.min=%d, time.max=%d\n",
		this.AvgSpearmanSim20, this.MinSpearmanSim20, this.MaxSpearmanSim20,
		this.AvgTime20, this.MinTime20, this.MaxTime20)

	ret += fmt.Sprintf("\tSpearmanSim40: avg=%.4f, min=%.4f, max=%.4f, time.avg=%d, time.min=%d, time.max=%d\n",
		this.AvgSpearmanSim40, this.MinSpearmanSim40, this.MaxSpearmanSim40,
		this.AvgTime40, this.MinTime40, this.MaxTime40)

	return ret
}
