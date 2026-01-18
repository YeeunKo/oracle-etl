// Package buffer는 IO 버퍼 설정 및 최적화를 제공합니다.
package buffer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	require.NotNil(t, config)

	// plan.md에 명시된 최적 버퍼 크기 검증
	assert.Equal(t, 1000, config.FetchArraySize, "Oracle fetch array size는 1000이어야 함")
	assert.Equal(t, 1000, config.PrefetchCount, "Oracle prefetch count는 1000이어야 함")
	assert.Equal(t, 64*1024, config.JSONLBufferSize, "JSONL 버퍼는 64KB여야 함")
	assert.Equal(t, 32*1024, config.GzipBufferSize, "Gzip 버퍼는 32KB여야 함")
	assert.Equal(t, 16*1024*1024, config.GCSChunkSize, "GCS chunk size는 16MB여야 함")
	assert.Equal(t, 10000, config.ChunkSize, "청크 row 수는 10000이어야 함")
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
		errorField  string
	}{
		{
			name:        "기본 설정 유효",
			config:      DefaultConfig(),
			expectError: false,
		},
		{
			name: "FetchArraySize 0 유효하지 않음",
			config: Config{
				FetchArraySize: 0,
				PrefetchCount:  1000,
				JSONLBufferSize: 64 * 1024,
				GzipBufferSize:  32 * 1024,
				GCSChunkSize:    16 * 1024 * 1024,
				ChunkSize:       10000,
			},
			expectError: true,
			errorField:  "FetchArraySize",
		},
		{
			name: "음수 버퍼 크기 유효하지 않음",
			config: Config{
				FetchArraySize: 1000,
				PrefetchCount:  1000,
				JSONLBufferSize: -1,
				GzipBufferSize:  32 * 1024,
				GCSChunkSize:    16 * 1024 * 1024,
				ChunkSize:       10000,
			},
			expectError: true,
			errorField:  "JSONLBufferSize",
		},
		{
			name: "GCS chunk size 너무 작음",
			config: Config{
				FetchArraySize: 1000,
				PrefetchCount:  1000,
				JSONLBufferSize: 64 * 1024,
				GzipBufferSize:  32 * 1024,
				GCSChunkSize:    1024, // 1KB - 너무 작음
				ChunkSize:       10000,
			},
			expectError: true,
			errorField:  "GCSChunkSize",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorField)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestConfig_HighPerformance(t *testing.T) {
	config := HighPerformanceConfig()

	require.NotNil(t, config)

	// 고성능 설정은 기본값보다 더 큰 버퍼 사용
	defaultConfig := DefaultConfig()

	assert.GreaterOrEqual(t, config.FetchArraySize, defaultConfig.FetchArraySize)
	assert.GreaterOrEqual(t, config.PrefetchCount, defaultConfig.PrefetchCount)
	assert.GreaterOrEqual(t, config.JSONLBufferSize, defaultConfig.JSONLBufferSize)
}

func TestConfig_LowMemory(t *testing.T) {
	config := LowMemoryConfig()

	require.NotNil(t, config)

	// 저메모리 설정은 기본값보다 작은 버퍼 사용
	defaultConfig := DefaultConfig()

	assert.LessOrEqual(t, config.FetchArraySize, defaultConfig.FetchArraySize)
	assert.LessOrEqual(t, config.JSONLBufferSize, defaultConfig.JSONLBufferSize)
}

func TestConfig_WithOptions(t *testing.T) {
	config := DefaultConfig()

	// 옵션 적용
	config = config.WithFetchArraySize(2000)
	assert.Equal(t, 2000, config.FetchArraySize)

	config = config.WithChunkSize(20000)
	assert.Equal(t, 20000, config.ChunkSize)

	config = config.WithJSONLBufferSize(128 * 1024)
	assert.Equal(t, 128*1024, config.JSONLBufferSize)
}

func TestConfig_EstimatedMemoryUsage(t *testing.T) {
	config := DefaultConfig()

	// 메모리 사용량 추정
	memUsage := config.EstimatedMemoryUsage()

	// 기본 설정에서 예상 메모리 사용량 (대략적인 범위 검증)
	assert.Greater(t, memUsage, int64(0))
	assert.Less(t, memUsage, int64(100*1024*1024)) // 100MB 미만
}

func TestConstants(t *testing.T) {
	// 상수 값 검증
	assert.Equal(t, 1000, OracleFetchArraySize)
	assert.Equal(t, 1000, OraclePrefetchCount)
	assert.Equal(t, 64*1024, JSONLBufferSize)
	assert.Equal(t, 32*1024, GzipBufferSize)
	assert.Equal(t, 16*1024*1024, GCSChunkSize)
	assert.Equal(t, 256*1024, GCSMinChunkSize)
	assert.Equal(t, 10000, DefaultChunkSize)
}
