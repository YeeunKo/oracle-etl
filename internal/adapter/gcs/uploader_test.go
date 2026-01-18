package gcs

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUploadProgress(t *testing.T) {
	progress := UploadProgress{
		BytesWritten: 1000,
		BytesTotal:   10000,
		RowsWritten:  100,
		StartTime:    time.Now().Add(-1 * time.Second),
		LastUpdate:   time.Now(),
	}

	// 진행률 계산
	percent := progress.Percent()
	assert.Equal(t, 10.0, percent)

	// 전송 속도 계산
	rate := progress.BytesPerSecond()
	assert.Greater(t, rate, 0.0)
}

func TestUploadProgress_UnknownTotal(t *testing.T) {
	progress := UploadProgress{
		BytesWritten: 1000,
		BytesTotal:   -1, // 알 수 없음
		RowsWritten:  100,
		StartTime:    time.Now(),
		LastUpdate:   time.Now(),
	}

	// 총량을 모르면 진행률은 -1
	percent := progress.Percent()
	assert.Equal(t, -1.0, percent)
}

func TestNewStreamingUploader(t *testing.T) {
	config := GCSConfig{
		ProjectID:  "test-project",
		BucketName: "test-bucket",
	}
	client := NewMockClient(config)

	uploader := NewStreamingUploader(client)
	require.NotNil(t, uploader)
}

func TestStreamingUploader_Upload(t *testing.T) {
	config := GCSConfig{
		ProjectID:  "test-project",
		BucketName: "test-bucket",
	}
	client := NewMockClient(config)
	uploader := NewStreamingUploader(client)

	ctx := context.Background()
	objectPath := "TRP-001/v001/VBRP.jsonl.gz"

	rows := []map[string]interface{}{
		{"id": 1, "name": "first"},
		{"id": 2, "name": "second"},
		{"id": 3, "name": "third"},
	}

	result, err := uploader.Upload(ctx, objectPath, rows, nil)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, int64(3), result.RowsWritten)
	assert.Greater(t, result.BytesWritten, int64(0))
	assert.Greater(t, result.Duration, time.Duration(0))
}

func TestStreamingUploader_UploadWithProgress(t *testing.T) {
	config := GCSConfig{
		ProjectID:  "test-project",
		BucketName: "test-bucket",
	}
	client := NewMockClient(config)
	uploader := NewStreamingUploader(client)

	ctx := context.Background()
	objectPath := "TRP-001/v001/VBRP.jsonl.gz"

	rows := make([]map[string]interface{}, 100)
	for i := 0; i < 100; i++ {
		rows[i] = map[string]interface{}{"id": i, "name": "row"}
	}

	var progressCalls int32
	callback := func(progress UploadProgress) {
		atomic.AddInt32(&progressCalls, 1)
		assert.Greater(t, progress.RowsWritten, int64(0))
	}

	result, err := uploader.Upload(ctx, objectPath, rows, callback)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, int64(100), result.RowsWritten)
	assert.Greater(t, atomic.LoadInt32(&progressCalls), int32(0))
}

func TestStreamingUploader_UploadStream(t *testing.T) {
	config := GCSConfig{
		ProjectID:  "test-project",
		BucketName: "test-bucket",
	}
	client := NewMockClient(config)
	uploader := NewStreamingUploader(client)

	ctx := context.Background()
	objectPath := "TRP-001/v001/VBRK.jsonl.gz"

	// Row 채널 생성
	rowChan := make(chan map[string]interface{}, 10)
	go func() {
		for i := 0; i < 50; i++ {
			rowChan <- map[string]interface{}{"id": i, "value": "test"}
		}
		close(rowChan)
	}()

	result, err := uploader.UploadStream(ctx, objectPath, rowChan, nil)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, int64(50), result.RowsWritten)
}

func TestStreamingUploader_ContextCancel(t *testing.T) {
	config := GCSConfig{
		ProjectID:  "test-project",
		BucketName: "test-bucket",
	}
	client := NewMockClient(config)
	uploader := NewStreamingUploader(client)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 즉시 취소

	objectPath := "TRP-001/v001/VBRP.jsonl.gz"
	rows := []map[string]interface{}{{"id": 1}}

	_, err := uploader.Upload(ctx, objectPath, rows, nil)
	require.Error(t, err)
}

func TestUploadResult_CompressionRatio(t *testing.T) {
	result := UploadResult{
		BytesWritten:  100,
		BytesOriginal: 1000,
		RowsWritten:   10,
		Duration:      time.Second,
	}

	ratio := result.CompressionRatio()
	assert.Equal(t, 0.1, ratio)
}

func TestUploadResult_Throughput(t *testing.T) {
	result := UploadResult{
		BytesWritten:  1024 * 1024, // 1MB
		BytesOriginal: 10 * 1024 * 1024,
		RowsWritten:   1000,
		Duration:      time.Second,
	}

	// rows/sec
	rowsPerSec := result.RowsPerSecond()
	assert.Equal(t, 1000.0, rowsPerSec)

	// MB/sec
	mbPerSec := result.MBPerSecond()
	assert.Equal(t, 1.0, mbPerSec)
}

func TestVerifyCompressedOutput(t *testing.T) {
	// Mock writer that captures data
	var buf bytes.Buffer
	mockClient := &capturingMockClient{
		MockClient: NewMockClient(GCSConfig{
			ProjectID:  "test-project",
			BucketName: "test-bucket",
		}).(*MockClient),
		buffer: &buf,
	}

	uploader := NewStreamingUploader(mockClient)

	ctx := context.Background()
	rows := []map[string]interface{}{
		{"MANDT": "800", "VBELN": "0090000001"},
		{"MANDT": "800", "VBELN": "0090000002"},
	}

	_, err := uploader.Upload(ctx, "test/path.jsonl.gz", rows, nil)
	require.NoError(t, err)

	// 압축된 데이터 검증
	reader, err := gzip.NewReader(&buf)
	require.NoError(t, err)
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSuffix(string(decompressed), "\n"), "\n")
	assert.Len(t, lines, 2)

	for _, line := range lines {
		var row map[string]interface{}
		require.NoError(t, json.Unmarshal([]byte(line), &row))
		assert.Contains(t, row, "MANDT")
		assert.Contains(t, row, "VBELN")
	}
}

// capturingMockClient는 데이터를 캡처하는 Mock 클라이언트입니다
type capturingMockClient struct {
	*MockClient
	buffer *bytes.Buffer
}

func (c *capturingMockClient) NewWriter(ctx context.Context, objectPath string) (io.WriteCloser, error) {
	return &capturingWriter{buffer: c.buffer}, nil
}

type capturingWriter struct {
	buffer *bytes.Buffer
}

func (w *capturingWriter) Write(p []byte) (n int, err error) {
	return w.buffer.Write(p)
}

func (w *capturingWriter) Close() error {
	return nil
}
