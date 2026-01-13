package models

import (
	"fmt"
	"math"
)

type TestConfigResult struct {
	TotalDocs int
	TotalTime int64
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

func NewTestConfigResult(totalDocs int) TestConfigResult {
	return TestConfigResult{
		TotalDocs:                    totalDocs,
		TotalTime:                    0,
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

func (this *TestConfigResult) Push10(spearman float64, elapsedMicro int64) {
	this.AvgSpearmanSim10 += spearman
	if spearman < this.MinSpearmanSim10 {
		this.MinSpearmanSim10 = spearman
	}
	if spearman > this.MaxSpearmanSim10 {
		this.MaxSpearmanSim10 = spearman
	}

	this.AvgTime10 += elapsedMicro
	if elapsedMicro < this.MinTime10 {
		this.MinTime10 = elapsedMicro
	}
	if elapsedMicro > this.MaxTime10 {
		this.MaxTime10 = elapsedMicro
	}
}

func (this *TestConfigResult) Push20(spearman float64, elapsedMicro int64) {
	this.AvgSpearmanSim20 += spearman
	if spearman < this.MinSpearmanSim20 {
		this.MinSpearmanSim20 = spearman
	}
	if spearman > this.MaxSpearmanSim20 {
		this.MaxSpearmanSim20 = spearman
	}

	this.AvgTime20 += elapsedMicro
	if elapsedMicro < this.MinTime20 {
		this.MinTime20 = elapsedMicro
	}
	if elapsedMicro > this.MaxTime20 {
		this.MaxTime20 = elapsedMicro
	}
}

func (this *TestConfigResult) Push40(spearman float64, elapsedMicro int64) {
	this.AvgSpearmanSim40 += spearman
	if spearman < this.MinSpearmanSim40 {
		this.MinSpearmanSim40 = spearman
	}
	if spearman > this.MaxSpearmanSim40 {
		this.MaxSpearmanSim40 = spearman
	}

	this.AvgTime40 += elapsedMicro
	if elapsedMicro < this.MinTime40 {
		this.MinTime40 = elapsedMicro
	}
	if elapsedMicro > this.MaxTime40 {
		this.MaxTime40 = elapsedMicro
	}
}

func (t *TestConfigResult) String() string {
	return fmt.Sprintf(
		"%d,%d,"+
			"%.4f,%.4f,%.4f,%d,%d,%d,"+
			"%.4f,%.4f,%.4f,%d,%d,%d,"+
			"%.4f,%.4f,%.4f,%d,%d,%d",
		t.TotalDocs,
		t.TotalTime,

		t.AvgSpearmanSim10/50,
		t.MinSpearmanSim10,
		t.MaxSpearmanSim10,
		t.AvgTime10/50,
		t.MinTime10,
		t.MaxTime10,

		t.AvgSpearmanSim20/50,
		t.MinSpearmanSim20,
		t.MaxSpearmanSim20,
		t.AvgTime20/50,
		t.MinTime20,
		t.MaxTime20,

		t.AvgSpearmanSim40/50,
		t.MinSpearmanSim40,
		t.MaxSpearmanSim40,
		t.AvgTime40/50,
		t.MinTime40,
		t.MaxTime40,
	)
}
