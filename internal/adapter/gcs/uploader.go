// Package gcs는 Google Cloud Storage 클라이언트 및 업로드 기능을 제공합니다.
package gcs

import (
	"context"
	"fmt"
	
	"sync/atomic"
	"time"

	"oracle-etl/pkg/compress"
	"oracle-etl/pkg/jsonl"
)

// UploadProgress는 업로드 진행 상황을 나타냅니다
type UploadProgress struct {
	BytesWritten int64     // 압축 후 전송된 바이트 수
	BytesTotal   int64     // 총 바이트 수 (-1이면 알 수 없음)
	RowsWritten  int64     // 기록된 row 수
	StartTime    time.Time // 업로드 시작 시간
	LastUpdate   time.Time // 마지막 업데이트 시간
}

// Percent는 진행률을 반환합니다 (0-100, 알 수 없으면 -1)
func (p UploadProgress) Percent() float64 {
	if p.BytesTotal <= 0 {
		return -1
	}
	return float64(p.BytesWritten) / float64(p.BytesTotal) * 100
}

// BytesPerSecond는 전송 속도를 반환합니다
func (p UploadProgress) BytesPerSecond() float64 {
	elapsed := time.Since(p.StartTime).Seconds()
	if elapsed <= 0 {
		return 0
	}
	return float64(p.BytesWritten) / elapsed
}

// RowsPerSecond는 row 처리 속도를 반환합니다
func (p UploadProgress) RowsPerSecond() float64 {
	elapsed := time.Since(p.StartTime).Seconds()
	if elapsed <= 0 {
		return 0
	}
	return float64(p.RowsWritten) / elapsed
}

// ProgressCallback는 진행률 콜백 함수 타입입니다
type ProgressCallback func(progress UploadProgress)

// UploadResult는 업로드 완료 결과를 나타냅니다
type UploadResult struct {
	BytesWritten  int64         // 압축 후 전송된 바이트 수
	BytesOriginal int64         // 압축 전 원본 바이트 수
	RowsWritten   int64         // 기록된 row 수
	Duration      time.Duration // 업로드 소요 시간
	ObjectPath    string        // GCS 객체 경로
}

// CompressionRatio는 압축률을 반환합니다 (0.0-1.0)
func (r UploadResult) CompressionRatio() float64 {
	if r.BytesOriginal <= 0 {
		return 0
	}
	return float64(r.BytesWritten) / float64(r.BytesOriginal)
}

// RowsPerSecond는 row 처리 속도를 반환합니다
func (r UploadResult) RowsPerSecond() float64 {
	if r.Duration <= 0 {
		return 0
	}
	return float64(r.RowsWritten) / r.Duration.Seconds()
}

// MBPerSecond는 전송 속도 (MB/s)를 반환합니다
func (r UploadResult) MBPerSecond() float64 {
	if r.Duration <= 0 {
		return 0
	}
	return float64(r.BytesWritten) / (1024 * 1024) / r.Duration.Seconds()
}

// Uploader는 GCS 스트리밍 업로드 인터페이스입니다
type Uploader interface {
	// Upload는 row 슬라이스를 GCS에 업로드합니다
	Upload(ctx context.Context, objectPath string, rows []map[string]interface{}, callback ProgressCallback) (*UploadResult, error)

	// UploadStream은 채널에서 row를 읽어 스트리밍 업로드합니다
	UploadStream(ctx context.Context, objectPath string, rowChan <-chan map[string]interface{}, callback ProgressCallback) (*UploadResult, error)
}

// StreamingUploader는 Uploader 인터페이스의 구현체입니다
type StreamingUploader struct {
	client            Client
	progressInterval  time.Duration // 진행률 콜백 호출 간격
}

// NewStreamingUploader는 새로운 스트리밍 업로더를 생성합니다
func NewStreamingUploader(client Client) Uploader {
	return &StreamingUploader{
		client:           client,
		progressInterval: 100 * time.Millisecond,
	}
}

// Upload는 row 슬라이스를 GCS에 업로드합니다
func (u *StreamingUploader) Upload(ctx context.Context, objectPath string, rows []map[string]interface{}, callback ProgressCallback) (*UploadResult, error) {
	startTime := time.Now()

	// 컨텍스트 취소 확인
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// GCS writer 생성
	gcsWriter, err := u.client.NewWriter(ctx, objectPath)
	if err != nil {
		return nil, fmt.Errorf("GCS writer 생성 실패: %w", err)
	}
	defer gcsWriter.Close()

	// 파이프라인: JSONL -> Gzip -> GCS
	gzipWriter := compress.NewGzipWriter(gcsWriter)
	jsonlEncoder := jsonl.NewEncoder(gzipWriter)

	var rowsWritten int64
	lastCallback := time.Now()

	for _, row := range rows {
		// 컨텍스트 취소 확인
		select {
		case <-ctx.Done():
			_ = gzipWriter.Close() // 에러 경로에서 정리
			return nil, ctx.Err()
		default:
		}

		if err := jsonlEncoder.Encode(row); err != nil {
			_ = gzipWriter.Close() // 에러 경로에서 정리
			return nil, fmt.Errorf("row 인코딩 실패: %w", err)
		}
		atomic.AddInt64(&rowsWritten, 1)

		// 진행률 콜백
		if callback != nil && time.Since(lastCallback) >= u.progressInterval {
			callback(UploadProgress{
				BytesWritten: gzipWriter.BytesWritten(),
				BytesTotal:   -1, // 스트리밍에서는 총량 알 수 없음
				RowsWritten:  atomic.LoadInt64(&rowsWritten),
				StartTime:    startTime,
				LastUpdate:   time.Now(),
			})
			lastCallback = time.Now()
		}
	}

	// JSONL 버퍼 플러시
	if err := jsonlEncoder.Flush(); err != nil {
		_ = gzipWriter.Close() // 에러 경로에서 정리
		return nil, fmt.Errorf("JSONL 플러시 실패: %w", err)
	}

	// Gzip 스트림 닫기
	if err := gzipWriter.Close(); err != nil {
		return nil, fmt.Errorf("gzip 스트림 닫기 실패: %w", err)
	}

	// 최종 진행률 콜백
	if callback != nil {
		callback(UploadProgress{
			BytesWritten: gzipWriter.BytesWritten(),
			BytesTotal:   gzipWriter.BytesWritten(),
			RowsWritten:  atomic.LoadInt64(&rowsWritten),
			StartTime:    startTime,
			LastUpdate:   time.Now(),
		})
	}

	return &UploadResult{
		BytesWritten:  gzipWriter.BytesWritten(),
		BytesOriginal: gzipWriter.BytesRead(),
		RowsWritten:   atomic.LoadInt64(&rowsWritten),
		Duration:      time.Since(startTime),
		ObjectPath:    objectPath,
	}, nil
}

// UploadStream은 채널에서 row를 읽어 스트리밍 업로드합니다
func (u *StreamingUploader) UploadStream(ctx context.Context, objectPath string, rowChan <-chan map[string]interface{}, callback ProgressCallback) (*UploadResult, error) {
	startTime := time.Now()

	// GCS writer 생성
	gcsWriter, err := u.client.NewWriter(ctx, objectPath)
	if err != nil {
		return nil, fmt.Errorf("GCS writer 생성 실패: %w", err)
	}
	defer gcsWriter.Close()

	// 파이프라인: JSONL -> Gzip -> GCS
	gzipWriter := compress.NewGzipWriter(gcsWriter)
	jsonlEncoder := jsonl.NewEncoder(gzipWriter)

	var rowsWritten int64
	lastCallback := time.Now()

	for {
		select {
		case <-ctx.Done():
			_ = gzipWriter.Close() // 에러 경로에서 정리
			return nil, ctx.Err()

		case row, ok := <-rowChan:
			if !ok {
				// 채널 닫힘 - 업로드 완료
				if err := jsonlEncoder.Flush(); err != nil {
					_ = gzipWriter.Close() // 에러 경로에서 정리
					return nil, fmt.Errorf("JSONL 플러시 실패: %w", err)
				}

				if err := gzipWriter.Close(); err != nil {
					return nil, fmt.Errorf("gzip 스트림 닫기 실패: %w", err)
				}

				// 최종 진행률 콜백
				if callback != nil {
					callback(UploadProgress{
						BytesWritten: gzipWriter.BytesWritten(),
						BytesTotal:   gzipWriter.BytesWritten(),
						RowsWritten:  atomic.LoadInt64(&rowsWritten),
						StartTime:    startTime,
						LastUpdate:   time.Now(),
					})
				}

				return &UploadResult{
					BytesWritten:  gzipWriter.BytesWritten(),
					BytesOriginal: gzipWriter.BytesRead(),
					RowsWritten:   atomic.LoadInt64(&rowsWritten),
					Duration:      time.Since(startTime),
					ObjectPath:    objectPath,
				}, nil
			}

			if err := jsonlEncoder.Encode(row); err != nil {
				_ = gzipWriter.Close() // 에러 경로에서 정리
				return nil, fmt.Errorf("row 인코딩 실패: %w", err)
			}
			atomic.AddInt64(&rowsWritten, 1)

			// 진행률 콜백
			if callback != nil && time.Since(lastCallback) >= u.progressInterval {
				callback(UploadProgress{
					BytesWritten: gzipWriter.BytesWritten(),
					BytesTotal:   -1,
					RowsWritten:  atomic.LoadInt64(&rowsWritten),
					StartTime:    startTime,
					LastUpdate:   time.Now(),
				})
				lastCallback = time.Now()
			}
		}
	}
}

// PipelineUploader는 Oracle 추출과 GCS 업로드를 연결하는 파이프라인입니다
type PipelineUploader struct {
	uploader Uploader
	client   Client
}

// NewPipelineUploader는 새로운 파이프라인 업로더를 생성합니다
func NewPipelineUploader(client Client) *PipelineUploader {
	return &PipelineUploader{
		uploader: NewStreamingUploader(client),
		client:   client,
	}
}

// UploadTable은 테이블 데이터를 GCS에 업로드합니다
func (p *PipelineUploader) UploadTable(ctx context.Context, transportID, jobVersion, tableName string, rows []map[string]interface{}, callback ProgressCallback) (*UploadResult, error) {
	objectPath := p.client.ObjectPath(transportID, jobVersion, tableName)
	return p.uploader.Upload(ctx, objectPath, rows, callback)
}

// UploadTableStream은 테이블 데이터를 스트리밍으로 GCS에 업로드합니다
func (p *PipelineUploader) UploadTableStream(ctx context.Context, transportID, jobVersion, tableName string, rowChan <-chan map[string]interface{}, callback ProgressCallback) (*UploadResult, error) {
	objectPath := p.client.ObjectPath(transportID, jobVersion, tableName)
	return p.uploader.UploadStream(ctx, objectPath, rowChan, callback)
}
