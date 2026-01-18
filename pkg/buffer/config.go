// Package buffer는 IO 버퍼 설정 및 최적화를 제공합니다.
// ETL 파이프라인의 성능 최적화를 위한 버퍼 크기 상수와 설정을 정의합니다.
package buffer

import (
	"errors"
	"fmt"
)

// Oracle 관련 버퍼 상수
// plan.md에 정의된 최적 값 기반
const (
	// OracleFetchArraySize는 Oracle에서 한 번에 fetch할 row 수입니다.
	// 라운드트립을 최소화하면서 메모리 효율성을 유지합니다.
	OracleFetchArraySize = 1000

	// OraclePrefetchCount는 Oracle prefetch 버퍼의 row 수입니다.
	OraclePrefetchCount = 1000
)

// IO 관련 버퍼 상수
const (
	// JSONLBufferSize는 JSONL 인코더의 버퍼 크기입니다 (64KB).
	// 작은 쓰기 작업을 버퍼링하여 IO 효율성을 높입니다.
	JSONLBufferSize = 64 * 1024

	// GzipBufferSize는 gzip 압축 버퍼 크기입니다 (32KB).
	GzipBufferSize = 32 * 1024
)

// GCS 관련 상수
const (
	// GCSChunkSize는 GCS resumable upload의 청크 크기입니다 (16MB).
	// GCS 권장 최소 크기이며, 재시작 시 재전송 오버헤드를 최소화합니다.
	GCSChunkSize = 16 * 1024 * 1024

	// GCSMinChunkSize는 GCS upload의 최소 청크 크기입니다 (256KB).
	GCSMinChunkSize = 256 * 1024
)

// ETL 관련 상수
const (
	// DefaultChunkSize는 스트리밍 시 청크당 row 수입니다.
	DefaultChunkSize = 10000

	// DefaultParallelism은 기본 병렬 처리 수입니다.
	DefaultParallelism = 4

	// MaxParallelism은 최대 병렬 처리 수입니다.
	MaxParallelism = 16
)

// Config는 버퍼 및 성능 관련 설정을 나타냅니다
type Config struct {
	// Oracle 설정
	FetchArraySize int // 한 번에 fetch할 row 수
	PrefetchCount  int // prefetch 버퍼 row 수

	// IO 버퍼 설정
	JSONLBufferSize int // JSONL 버퍼 크기 (bytes)
	GzipBufferSize  int // Gzip 버퍼 크기 (bytes)

	// GCS 설정
	GCSChunkSize int // GCS 업로드 청크 크기 (bytes)

	// ETL 설정
	ChunkSize int // 스트리밍 청크 row 수
}

// DefaultConfig는 기본 설정을 반환합니다
// plan.md에 정의된 최적화된 값을 사용합니다
func DefaultConfig() Config {
	return Config{
		FetchArraySize:  OracleFetchArraySize,
		PrefetchCount:   OraclePrefetchCount,
		JSONLBufferSize: JSONLBufferSize,
		GzipBufferSize:  GzipBufferSize,
		GCSChunkSize:    GCSChunkSize,
		ChunkSize:       DefaultChunkSize,
	}
}

// HighPerformanceConfig는 고성능 설정을 반환합니다
// 더 큰 버퍼를 사용하여 처리량을 극대화합니다
func HighPerformanceConfig() Config {
	return Config{
		FetchArraySize:  2000,              // 더 큰 배치
		PrefetchCount:   2000,              // 더 많은 prefetch
		JSONLBufferSize: 128 * 1024,        // 128KB
		GzipBufferSize:  64 * 1024,         // 64KB
		GCSChunkSize:    32 * 1024 * 1024,  // 32MB
		ChunkSize:       20000,             // 더 큰 청크
	}
}

// LowMemoryConfig는 저메모리 설정을 반환합니다
// 메모리 사용량을 최소화하면서 안정적인 처리를 보장합니다
func LowMemoryConfig() Config {
	return Config{
		FetchArraySize:  500,               // 작은 배치
		PrefetchCount:   500,               // 적은 prefetch
		JSONLBufferSize: 32 * 1024,         // 32KB
		GzipBufferSize:  16 * 1024,         // 16KB
		GCSChunkSize:    8 * 1024 * 1024,   // 8MB
		ChunkSize:       5000,              // 작은 청크
	}
}

// Validate는 설정의 유효성을 검사합니다
func (c Config) Validate() error {
	if c.FetchArraySize <= 0 {
		return errors.New("FetchArraySize는 양수여야 합니다")
	}
	if c.PrefetchCount <= 0 {
		return errors.New("PrefetchCount는 양수여야 합니다")
	}
	if c.JSONLBufferSize <= 0 {
		return errors.New("JSONLBufferSize는 양수여야 합니다")
	}
	if c.GzipBufferSize <= 0 {
		return errors.New("GzipBufferSize는 양수여야 합니다")
	}
	if c.GCSChunkSize < GCSMinChunkSize {
		return fmt.Errorf("GCSChunkSize는 최소 %d bytes여야 합니다", GCSMinChunkSize)
	}
	if c.ChunkSize <= 0 {
		return errors.New("ChunkSize는 양수여야 합니다")
	}
	return nil
}

// WithFetchArraySize는 FetchArraySize를 설정한 새 Config를 반환합니다
func (c Config) WithFetchArraySize(size int) Config {
	c.FetchArraySize = size
	return c
}

// WithPrefetchCount는 PrefetchCount를 설정한 새 Config를 반환합니다
func (c Config) WithPrefetchCount(count int) Config {
	c.PrefetchCount = count
	return c
}

// WithJSONLBufferSize는 JSONLBufferSize를 설정한 새 Config를 반환합니다
func (c Config) WithJSONLBufferSize(size int) Config {
	c.JSONLBufferSize = size
	return c
}

// WithGzipBufferSize는 GzipBufferSize를 설정한 새 Config를 반환합니다
func (c Config) WithGzipBufferSize(size int) Config {
	c.GzipBufferSize = size
	return c
}

// WithGCSChunkSize는 GCSChunkSize를 설정한 새 Config를 반환합니다
func (c Config) WithGCSChunkSize(size int) Config {
	c.GCSChunkSize = size
	return c
}

// WithChunkSize는 ChunkSize를 설정한 새 Config를 반환합니다
func (c Config) WithChunkSize(size int) Config {
	c.ChunkSize = size
	return c
}

// EstimatedMemoryUsage는 이 설정으로 예상되는 메모리 사용량을 반환합니다
// 단위: bytes
func (c Config) EstimatedMemoryUsage() int64 {
	// 대략적인 메모리 사용량 추정
	// - FetchArraySize * 평균 row 크기 (약 1KB)
	// - JSONL 버퍼
	// - Gzip 버퍼
	// - GCS 업로드 버퍼

	avgRowSize := int64(1024) // 1KB per row 추정

	oracleBuffer := int64(c.FetchArraySize) * avgRowSize
	jsonlBuffer := int64(c.JSONLBufferSize)
	gzipBuffer := int64(c.GzipBufferSize)
	gcsBuffer := int64(c.GCSChunkSize)

	// 안전 마진 (2x)
	total := (oracleBuffer + jsonlBuffer + gzipBuffer + gcsBuffer) * 2

	return total
}

// String은 설정의 문자열 표현을 반환합니다
func (c Config) String() string {
	return fmt.Sprintf(
		"Config{FetchArraySize: %d, PrefetchCount: %d, JSONLBufferSize: %dKB, GzipBufferSize: %dKB, GCSChunkSize: %dMB, ChunkSize: %d}",
		c.FetchArraySize,
		c.PrefetchCount,
		c.JSONLBufferSize/1024,
		c.GzipBufferSize/1024,
		c.GCSChunkSize/(1024*1024),
		c.ChunkSize,
	)
}
