package core

import "testing"

func TestRun(t *testing.T) {
	Run("BV1cv4y1N7vP", "onedrive")
}

func TestFormatConversion(t *testing.T) {
	formatConversion("tes")
}

func BenchmarkName(b *testing.B) {
	for i := 0; i < b.N; i++ {
		formatConversion("tes")
	}
}
