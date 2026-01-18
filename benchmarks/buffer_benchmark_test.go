// Package benchmarks는 성능 벤치마크 테스트를 제공합니다.
package benchmarks

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"testing"

	"oracle-etl/pkg/buffer"
)

// BenchmarkBufferConfigs는 다양한 버퍼 설정의 메모리 사용량을 비교합니다
func BenchmarkBufferConfigs(b *testing.B) {
	configs := map[string]buffer.Config{
		"Default":         buffer.DefaultConfig(),
		"HighPerformance": buffer.HighPerformanceConfig(),
		"LowMemory":       buffer.LowMemoryConfig(),
	}

	for name, cfg := range configs {
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				// 메모리 사용량 추정
				_ = cfg.EstimatedMemoryUsage()
			}
		})
	}
}

// BenchmarkJSONLBuffer는 JSONL 버퍼 크기별 성능을 측정합니다
func BenchmarkJSONLBuffer(b *testing.B) {
	sizes := []int{32 * 1024, 64 * 1024, 128 * 1024, 256 * 1024}
	
	// 샘플 데이터
	sampleData := map[string]interface{}{
		"MANDT":  "800",
		"VBELN":  "0090000001",
		"POSNR":  float64(10),
		"MATNR":  "MAT-12345678",
		"NETWR":  1234.56,
		"WAERK":  "USD",
		"ERDAT":  "2024-01-15",
		"ERZET":  "10:30:45",
		"ERNAM":  "USER001",
		"KWMENG": 100.0,
	}

	for _, size := range sizes {
		b.Run(
			bufferSizeName(size),
			func(b *testing.B) {
				buf := bytes.NewBuffer(make([]byte, 0, size))
				encoder := json.NewEncoder(buf)

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					buf.Reset()
					for j := 0; j < 1000; j++ {
						_ = encoder.Encode(sampleData)
					}
				}
				b.ReportMetric(float64(buf.Len())/1024, "KB/batch")
			},
		)
	}
}

// BenchmarkGzipBuffer는 Gzip 버퍼 크기별 압축 성능을 측정합니다
func BenchmarkGzipBuffer(b *testing.B) {
	sizes := []int{16 * 1024, 32 * 1024, 64 * 1024}
	
	// 테스트 데이터 생성 (1MB)
	testData := bytes.Repeat([]byte(`{"MANDT":"800","VBELN":"0090000001","POSNR":10,"MATNR":"MAT-12345678"}`), 10000)

	for _, size := range sizes {
		b.Run(
			gzipBufferName(size),
			func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					var buf bytes.Buffer
					gw, _ := gzip.NewWriterLevel(&buf, gzip.BestSpeed)
					_, _ = gw.Write(testData)
					gw.Close()
				}
			},
		)
	}
}

// BenchmarkFetchArraySize는 다양한 fetch array 크기에서 시뮬레이션된 성능을 측정합니다
func BenchmarkFetchArraySize(b *testing.B) {
	sizes := []int{500, 1000, 2000, 5000}
	
	// 시뮬레이션된 row 데이터
	rowData := map[string]interface{}{
		"col1": "value1",
		"col2": float64(12345),
		"col3": "2024-01-15T10:30:45Z",
	}

	for _, size := range sizes {
		b.Run(
			fetchArrayName(size),
			func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					// fetch 시뮬레이션
					batch := make([]map[string]interface{}, 0, size)
					for j := 0; j < size; j++ {
						batch = append(batch, rowData)
					_ = batch // 벤치마크 측정용
					}
				}
			},
		)
	}
}

func bufferSizeName(size int) string {
	return itoa(size/1024) + "KB"
}

func gzipBufferName(size int) string {
	return "gzip_" + itoa(size/1024) + "KB"
}

func fetchArrayName(size int) string {
	return "fetch_" + itoa(size)
}
