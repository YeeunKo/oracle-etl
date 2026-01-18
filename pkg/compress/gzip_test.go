// Package compress는 데이터 압축 기능을 제공합니다.
package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGzipWriter_Write(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "간단한 문자열 압축",
			input: "Hello, World!",
		},
		{
			name:  "한글 문자열 압축",
			input: "안녕하세요, 세계!",
		},
		{
			name:  "반복 패턴 (높은 압축률)",
			input: strings.Repeat("ABCDEFGH", 1000),
		},
		{
			name:  "빈 문자열",
			input: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			writer := NewGzipWriter(&buf)

			n, err := writer.Write([]byte(tt.input))
			require.NoError(t, err)
			assert.Equal(t, len(tt.input), n)

			err = writer.Close()
			require.NoError(t, err)

			// 압축된 데이터 해제하여 검증
			reader, err := gzip.NewReader(&buf)
			require.NoError(t, err)

			decompressed, err := io.ReadAll(reader)
			require.NoError(t, err)
			require.NoError(t, reader.Close())

			assert.Equal(t, tt.input, string(decompressed))
		})
	}
}

func TestGzipWriter_CompressionRatio(t *testing.T) {
	// 반복 패턴이 많은 데이터는 높은 압축률을 보여야 함
	input := strings.Repeat("Oracle ETL Pipeline Data Row ", 10000)
	inputBytes := []byte(input)

	var buf bytes.Buffer
	writer := NewGzipWriter(&buf)

	_, err := writer.Write(inputBytes)
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	compressedSize := buf.Len()
	originalSize := len(inputBytes)

	// 압축률 확인 (반복 패턴은 매우 잘 압축됨)
	compressionRatio := float64(compressedSize) / float64(originalSize)
	t.Logf("원본 크기: %d, 압축 크기: %d, 압축률: %.2f%%", originalSize, compressedSize, compressionRatio*100)

	// 반복 패턴은 최소 90% 이상 압축되어야 함
	assert.Less(t, compressionRatio, 0.1, "반복 패턴은 높은 압축률을 보여야 함")
}

func TestGzipWriter_MultipleWrites(t *testing.T) {
	var buf bytes.Buffer
	writer := NewGzipWriter(&buf)

	// 여러 번 쓰기
	chunks := []string{
		`{"id": 1, "name": "first"}`,
		`{"id": 2, "name": "second"}`,
		`{"id": 3, "name": "third"}`,
	}

	for _, chunk := range chunks {
		_, err := writer.Write([]byte(chunk + "\n"))
		require.NoError(t, err)
	}

	err := writer.Close()
	require.NoError(t, err)

	// 압축 해제하여 검증
	reader, err := gzip.NewReader(&buf)
	require.NoError(t, err)

	decompressed, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.NoError(t, reader.Close())

	expected := strings.Join(chunks, "\n") + "\n"
	assert.Equal(t, expected, string(decompressed))
}

func TestGzipWriter_Flush(t *testing.T) {
	var buf bytes.Buffer
	writer := NewGzipWriter(&buf)

	_, err := writer.Write([]byte("test data"))
	require.NoError(t, err)

	// Flush 호출
	err = writer.Flush()
	require.NoError(t, err)

	// Flush 후에도 추가 쓰기 가능해야 함
	_, err = writer.Write([]byte(" more data"))
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	// 압축 해제하여 검증
	reader, err := gzip.NewReader(&buf)
	require.NoError(t, err)

	decompressed, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.NoError(t, reader.Close())

	assert.Equal(t, "test data more data", string(decompressed))
}

func TestGzipWriter_BytesWritten(t *testing.T) {
	var buf bytes.Buffer
	writer := NewGzipWriter(&buf)

	input := []byte("test data for counting")
	_, err := writer.Write(input)
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	bytesWritten := writer.BytesWritten()
	assert.Greater(t, bytesWritten, int64(0))
	assert.Equal(t, int64(buf.Len()), bytesWritten)
}

func TestGzipWriter_BytesRead(t *testing.T) {
	var buf bytes.Buffer
	writer := NewGzipWriter(&buf)

	input := []byte("test data for counting input")
	_, err := writer.Write(input)
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	bytesRead := writer.BytesRead()
	assert.Equal(t, int64(len(input)), bytesRead)
}

func TestGzipWriter_BestCompression(t *testing.T) {
	input := strings.Repeat("Compression Test Data ", 1000)
	inputBytes := []byte(input)

	// 기본 압축
	var bufDefault bytes.Buffer
	writerDefault := NewGzipWriter(&bufDefault)
	_, _ = writerDefault.Write(inputBytes)
	writerDefault.Close()

	// 최대 압축
	var bufBest bytes.Buffer
	writerBest := NewGzipWriterLevel(&bufBest, gzip.BestCompression)
	_, _ = writerBest.Write(inputBytes)
	writerBest.Close()

	// 최대 압축이 더 작거나 같아야 함
	assert.LessOrEqual(t, bufBest.Len(), bufDefault.Len())
}

func TestGzipWriter_LargeData(t *testing.T) {
	// 1MB 데이터 생성
	input := make([]byte, 1024*1024)
	for i := range input {
		input[i] = byte(i % 256)
	}

	var buf bytes.Buffer
	writer := NewGzipWriter(&buf)

	_, err := writer.Write(input)
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	// 압축 해제하여 검증
	reader, err := gzip.NewReader(&buf)
	require.NoError(t, err)

	decompressed, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.NoError(t, reader.Close())

	assert.Equal(t, input, decompressed)
}

func BenchmarkGzipWriter_Write(b *testing.B) {
	input := []byte(strings.Repeat("Oracle ETL Pipeline Benchmark Data ", 100))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		writer := NewGzipWriter(&buf)
		_, _ = writer.Write(input)
		writer.Close()
	}
}

func BenchmarkGzipWriter_LargeData(b *testing.B) {
	// 1MB 데이터
	input := make([]byte, 1024*1024)
	for i := range input {
		input[i] = byte(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		writer := NewGzipWriter(&buf)
		_, _ = writer.Write(input)
		writer.Close()
	}
}
