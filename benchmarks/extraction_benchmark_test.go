// Package benchmarks는 성능 벤치마크 테스트를 제공합니다.
package benchmarks

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"testing"

	"oracle-etl/pkg/buffer"
)

// sampleRow는 벤치마크용 샘플 row 데이터입니다
var sampleRow = map[string]interface{}{
	"MANDT":    "800",
	"VBELN":    "0090000001",
	"POSNR":    float64(10),
	"MATNR":    "MAT-1234567890",
	"ARKTX":    "Sample Material Description Text",
	"NETWR":    1234.56,
	"WAERK":    "USD",
	"FKIMG":    100.0,
	"VRKME":    "EA",
	"ERDAT":    "2024-01-15",
	"ERZET":    "10:30:45",
	"ERNAM":    "USER001",
	"AEDAT":    "2024-01-16",
	"AESSION":  "SESSION123",
	"KWMENG":   100.0,
	"KZWI1":    50.0,
	"KZWI2":    25.0,
	"MWSBP":    12.5,
	"NETPR":    12.3456,
	"PRSDT":    "2024-01-15",
}

// BenchmarkFullPipeline_JSONL은 JSONL 인코딩 성능을 측정합니다
func BenchmarkFullPipeline_JSONL(b *testing.B) {
	rowCounts := []int{1000, 10000, 100000}

	for _, count := range rowCounts {
		b.Run(
			"rows="+itoa(count),
			func(b *testing.B) {
				rows := make([]map[string]interface{}, count)
				for i := 0; i < count; i++ {
					rows[i] = sampleRow
				}

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					var buf bytes.Buffer
					buf.Grow(buffer.JSONLBufferSize)
					encoder := json.NewEncoder(&buf)

					for _, row := range rows {
						_ = encoder.Encode(row)
					}
				}

				b.ReportMetric(float64(count)/b.Elapsed().Seconds(), "rows/sec")
			},
		)
	}
}

// BenchmarkFullPipeline_JSONLWithGzip은 JSONL + Gzip 압축 성능을 측정합니다
func BenchmarkFullPipeline_JSONLWithGzip(b *testing.B) {
	rowCounts := []int{1000, 10000, 100000}

	for _, count := range rowCounts {
		b.Run(
			"rows="+itoa(count),
			func(b *testing.B) {
				rows := make([]map[string]interface{}, count)
				for i := 0; i < count; i++ {
					rows[i] = sampleRow
				}

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					var buf bytes.Buffer
					buf.Grow(buffer.GzipBufferSize)
					
					gw, _ := gzip.NewWriterLevel(&buf, gzip.BestSpeed)
					encoder := json.NewEncoder(gw)

					for _, row := range rows {
						_ = encoder.Encode(row)
					}
					gw.Close()
				}

				b.ReportMetric(float64(count)/b.Elapsed().Seconds(), "rows/sec")
			},
		)
	}
}

// BenchmarkFullPipeline_MemoryAllocation은 메모리 할당 패턴을 측정합니다
func BenchmarkFullPipeline_MemoryAllocation(b *testing.B) {
	configs := map[string]buffer.Config{
		"Default":         buffer.DefaultConfig(),
		"HighPerformance": buffer.HighPerformanceConfig(),
		"LowMemory":       buffer.LowMemoryConfig(),
	}

	rowCount := 10000

	for name, cfg := range configs {
		b.Run(name, func(b *testing.B) {
			rows := make([]map[string]interface{}, rowCount)
			for i := 0; i < rowCount; i++ {
				rows[i] = sampleRow
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				// JSONL 버퍼 할당
				jsonlBuf := bytes.NewBuffer(make([]byte, 0, cfg.JSONLBufferSize))
				
				// Gzip 버퍼 할당
				var gzipBuf bytes.Buffer
				gzipBuf.Grow(cfg.GzipBufferSize)
				
				gw, _ := gzip.NewWriterLevel(&gzipBuf, gzip.BestSpeed)
				encoder := json.NewEncoder(jsonlBuf)

				// 청크 단위 처리
				chunkSize := cfg.ChunkSize
				for start := 0; start < len(rows); start += chunkSize {
					end := start + chunkSize
					if end > len(rows) {
						end = len(rows)
					}

					jsonlBuf.Reset()
					for _, row := range rows[start:end] {
						_ = encoder.Encode(row)
					}

					_, _ = gw.Write(jsonlBuf.Bytes())
				}
				gw.Close()
			}
		})
	}
}

// BenchmarkChunkProcessing은 청크 단위 처리 성능을 측정합니다
func BenchmarkChunkProcessing(b *testing.B) {
	chunkSizes := []int{1000, 5000, 10000, 20000}
	totalRows := 100000

	for _, chunkSize := range chunkSizes {
		b.Run(
			"chunk="+itoa(chunkSize),
			func(b *testing.B) {
				rows := make([]map[string]interface{}, totalRows)
				for i := 0; i < totalRows; i++ {
					rows[i] = sampleRow
				}

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					var buf bytes.Buffer
					buf.Grow(buffer.JSONLBufferSize)
					encoder := json.NewEncoder(&buf)

					for start := 0; start < len(rows); start += chunkSize {
						end := start + chunkSize
						if end > len(rows) {
							end = len(rows)
						}

						buf.Reset()
						for _, row := range rows[start:end] {
							_ = encoder.Encode(row)
						}
					}
				}

				b.ReportMetric(float64(totalRows)/b.Elapsed().Seconds(), "rows/sec")
			},
		)
	}
}

// BenchmarkCompressionRatio는 압축률과 처리 속도의 트레이드오프를 측정합니다
func BenchmarkCompressionRatio(b *testing.B) {
	levels := map[string]int{
		"BestSpeed":       gzip.BestSpeed,
		"DefaultLevel":    gzip.DefaultCompression,
		"BestCompression": gzip.BestCompression,
	}

	rowCount := 10000
	rows := make([]map[string]interface{}, rowCount)
	for i := 0; i < rowCount; i++ {
		rows[i] = sampleRow
	}

	// JSONL 데이터 생성
	var jsonlBuf bytes.Buffer
	encoder := json.NewEncoder(&jsonlBuf)
	for _, row := range rows {
		_ = encoder.Encode(row)
	}
	jsonlData := jsonlBuf.Bytes()

	for name, level := range levels {
		b.Run(name, func(b *testing.B) {
			b.ResetTimer()
			var totalCompressed int64

			for i := 0; i < b.N; i++ {
				var buf bytes.Buffer
				gw, _ := gzip.NewWriterLevel(&buf, level)
				_, _ = gw.Write(jsonlData)
				gw.Close()
				totalCompressed += int64(buf.Len())
			}

			avgCompressed := totalCompressed / int64(b.N)
			ratio := float64(len(jsonlData)) / float64(avgCompressed)
			
			b.ReportMetric(ratio, "compression_ratio")
			b.ReportMetric(float64(rowCount)/b.Elapsed().Seconds(), "rows/sec")
		})
	}
}
