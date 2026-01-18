// Package gcs는 Google Cloud Storage 클라이언트 및 업로드 기능을 제공합니다.
// 스트리밍 업로드, resumable 업로드, 진행률 추적을 지원합니다.
package gcs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

// 기본 상수 정의
const (
	// DefaultChunkSize는 resumable 업로드의 기본 청크 크기입니다 (16MB)
	DefaultChunkSize = 16 * 1024 * 1024

	// DefaultTimeout은 GCS 작업의 기본 타임아웃입니다
	DefaultTimeout = 10 * time.Minute
)

// GCSConfig는 GCS 클라이언트 설정을 정의합니다
type GCSConfig struct {
	ProjectID       string        `mapstructure:"project_id"`       // GCP 프로젝트 ID
	BucketName      string        `mapstructure:"bucket_name"`      // 대상 버킷 이름
	CredentialsFile string        `mapstructure:"credentials_file"` // 서비스 계정 JSON 파일 경로
	ChunkSize       int           `mapstructure:"chunk_size"`       // resumable 업로드 청크 크기
	Timeout         time.Duration `mapstructure:"timeout"`          // 작업 타임아웃
}

// Validate는 설정의 유효성을 검사합니다
func (c *GCSConfig) Validate() error {
	if c.ProjectID == "" {
		return errors.New("GCS ProjectID가 설정되지 않음")
	}
	if c.BucketName == "" {
		return errors.New("GCS BucketName이 설정되지 않음")
	}
	return nil
}

// ApplyDefaults는 기본값을 적용합니다
func (c *GCSConfig) ApplyDefaults() {
	if c.ChunkSize <= 0 {
		c.ChunkSize = DefaultChunkSize
	}
	if c.Timeout <= 0 {
		c.Timeout = DefaultTimeout
	}
}

// Client는 GCS 클라이언트 인터페이스입니다
type Client interface {
	// Ping은 GCS 연결을 테스트합니다
	Ping(ctx context.Context) error

	// NewWriter는 객체에 쓰기 위한 Writer를 생성합니다
	NewWriter(ctx context.Context, objectPath string) (io.WriteCloser, error)

	// ObjectPath는 표준 객체 경로를 생성합니다
	ObjectPath(transportID, jobVersion, tableName string) string

	// FullGCSPath는 전체 GCS URI를 반환합니다
	FullGCSPath(transportID, jobVersion, tableName string) string

	// BucketName은 버킷 이름을 반환합니다
	BucketName() string

	// Close는 클라이언트를 닫습니다
	Close() error
}

// gcsClient는 실제 GCS 클라이언트 구현체입니다
type gcsClient struct {
	client     *storage.Client
	bucket     *storage.BucketHandle
	config     GCSConfig
}

// NewClient는 새로운 GCS 클라이언트를 생성합니다
func NewClient(ctx context.Context, config GCSConfig) (Client, error) {
	config.ApplyDefaults()

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("GCS 설정 유효성 검사 실패: %w", err)
	}

	var opts []option.ClientOption
	if config.CredentialsFile != "" {
		opts = append(opts, option.WithCredentialsFile(config.CredentialsFile))
	}

	client, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("GCS 클라이언트 생성 실패: %w", err)
	}

	bucket := client.Bucket(config.BucketName)

	return &gcsClient{
		client: client,
		bucket: bucket,
		config: config,
	}, nil
}

// Ping은 GCS 연결을 테스트합니다
func (c *gcsClient) Ping(ctx context.Context) error {
	_, err := c.bucket.Attrs(ctx)
	if err != nil {
		return fmt.Errorf("GCS 버킷 접근 실패: %w", err)
	}
	return nil
}

// NewWriter는 객체에 쓰기 위한 Writer를 생성합니다
func (c *gcsClient) NewWriter(ctx context.Context, objectPath string) (io.WriteCloser, error) {
	obj := c.bucket.Object(objectPath)
	writer := obj.NewWriter(ctx)

	// Resumable 업로드를 위한 청크 크기 설정
	writer.ChunkSize = c.config.ChunkSize

	// Content-Type 설정
	writer.ContentType = "application/gzip"
	writer.ContentEncoding = "gzip"

	return writer, nil
}

// ObjectPath는 표준 객체 경로를 생성합니다
// 패턴: {transport_id}/{job_version}/{table_name}.jsonl.gz
func (c *gcsClient) ObjectPath(transportID, jobVersion, tableName string) string {
	return fmt.Sprintf("%s/%s/%s.jsonl.gz", transportID, jobVersion, tableName)
}

// FullGCSPath는 전체 GCS URI를 반환합니다
// 패턴: gs://{bucket}/{transport_id}/{job_version}/{table_name}.jsonl.gz
func (c *gcsClient) FullGCSPath(transportID, jobVersion, tableName string) string {
	return fmt.Sprintf("gs://%s/%s", c.config.BucketName, c.ObjectPath(transportID, jobVersion, tableName))
}

// BucketName은 버킷 이름을 반환합니다
func (c *gcsClient) BucketName() string {
	return c.config.BucketName
}

// Close는 클라이언트를 닫습니다
func (c *gcsClient) Close() error {
	return c.client.Close()
}

// MockClient는 테스트용 Mock GCS 클라이언트입니다
type MockClient struct {
	config GCSConfig
	closed bool
}

// NewMockClient는 테스트용 Mock 클라이언트를 생성합니다
func NewMockClient(config GCSConfig) Client {
	config.ApplyDefaults()
	return &MockClient{
		config: config,
		closed: false,
	}
}

// Ping은 Mock 연결 테스트입니다 (항상 성공)
func (m *MockClient) Ping(ctx context.Context) error {
	return nil
}

// NewWriter는 Mock Writer를 반환합니다
func (m *MockClient) NewWriter(ctx context.Context, objectPath string) (io.WriteCloser, error) {
	return &mockWriter{}, nil
}

// ObjectPath는 표준 객체 경로를 생성합니다
func (m *MockClient) ObjectPath(transportID, jobVersion, tableName string) string {
	return fmt.Sprintf("%s/%s/%s.jsonl.gz", transportID, jobVersion, tableName)
}

// FullGCSPath는 전체 GCS URI를 반환합니다
func (m *MockClient) FullGCSPath(transportID, jobVersion, tableName string) string {
	return fmt.Sprintf("gs://%s/%s", m.config.BucketName, m.ObjectPath(transportID, jobVersion, tableName))
}

// BucketName은 버킷 이름을 반환합니다
func (m *MockClient) BucketName() string {
	return m.config.BucketName
}

// Close는 Mock 클라이언트를 닫습니다
func (m *MockClient) Close() error {
	m.closed = true
	return nil
}

// mockWriter는 테스트용 Mock io.WriteCloser입니다
type mockWriter struct {
	bytesWritten int64
}

func (w *mockWriter) Write(p []byte) (n int, err error) {
	n = len(p)
	w.bytesWritten += int64(n)
	return n, nil
}

func (w *mockWriter) Close() error {
	return nil
}
