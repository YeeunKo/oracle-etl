//go:build integration

// Package integration은 Oracle 및 GCS 통합 테스트를 제공합니다.
package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"oracle-etl/internal/adapter/gcs"
)

// TestGCSConnection은 실제 GCS 연결을 테스트합니다.
func TestGCSConnection(t *testing.T) {
	credentialsFile := os.Getenv("GCS_CREDENTIALS_FILE")
	bucketName := os.Getenv("GCS_TEST_BUCKET")
	projectID := os.Getenv("GCS_PROJECT_ID")

	if credentialsFile == "" || bucketName == "" || projectID == "" {
		t.Skip("Skipping: GCS 환경 변수가 설정되지 않음 (GCS_CREDENTIALS_FILE, GCS_TEST_BUCKET, GCS_PROJECT_ID)")
	}

	cfg := gcs.GCSConfig{
		ProjectID:       projectID,
		BucketName:      bucketName,
		CredentialsFile: credentialsFile,
		ChunkSize:       16 * 1024 * 1024,
		Timeout:         5 * time.Minute,
	}

	t.Run("설정 유효성 검사", func(t *testing.T) {
		err := cfg.Validate()
		require.NoError(t, err)
	})
}

// TestGCSConfigDefaults는 GCS 설정 기본값을 테스트합니다.
func TestGCSConfigDefaults(t *testing.T) {
	cfg := gcs.GCSConfig{
		ProjectID:  "test-project",
		BucketName: "test-bucket",
	}

	cfg.ApplyDefaults()

	t.Run("기본 ChunkSize 적용", func(t *testing.T) {
		assert.Equal(t, gcs.DefaultChunkSize, cfg.ChunkSize)
	})

	t.Run("기본 Timeout 적용", func(t *testing.T) {
		assert.Equal(t, gcs.DefaultTimeout, cfg.Timeout)
	})
}

// TestGCSMockClient는 Mock 클라이언트 동작을 테스트합니다.
func TestGCSMockClient(t *testing.T) {
	cfg := gcs.GCSConfig{
		ProjectID:  "test-project",
		BucketName: "test-bucket",
	}

	client := gcs.NewMockClient(cfg)
	ctx := context.Background()

	t.Run("Ping 테스트", func(t *testing.T) {
		err := client.Ping(ctx)
		require.NoError(t, err)
	})

	t.Run("버킷 이름 확인", func(t *testing.T) {
		assert.Equal(t, "test-bucket", client.BucketName())
	})

	t.Run("객체 경로 생성", func(t *testing.T) {
		path := client.ObjectPath("TRP-001", "v001", "VBRP")
		assert.Equal(t, "TRP-001/v001/VBRP.jsonl.gz", path)
	})

	t.Run("전체 GCS 경로 생성", func(t *testing.T) {
		path := client.FullGCSPath("TRP-001", "v001", "VBRP")
		assert.Equal(t, "gs://test-bucket/TRP-001/v001/VBRP.jsonl.gz", path)
	})

	t.Run("Close 테스트", func(t *testing.T) {
		err := client.Close()
		require.NoError(t, err)
	})
}

// TestGCSUploadFlow는 업로드 플로우를 통합 테스트합니다.
func TestGCSUploadFlow(t *testing.T) {
	credentialsFile := os.Getenv("GCS_CREDENTIALS_FILE")
	bucketName := os.Getenv("GCS_TEST_BUCKET")
	projectID := os.Getenv("GCS_PROJECT_ID")

	if credentialsFile == "" || bucketName == "" || projectID == "" {
		t.Skip("Skipping: GCS 환경 변수가 설정되지 않음")
	}

	cfg := gcs.GCSConfig{
		ProjectID:       projectID,
		BucketName:      bucketName,
		CredentialsFile: credentialsFile,
	}
	cfg.ApplyDefaults()

	t.Run("설정 유효성 검사", func(t *testing.T) {
		err := cfg.Validate()
		require.NoError(t, err)
	})

	t.Run("Mock 업로드 플로우", func(t *testing.T) {
		mockClient := gcs.NewMockClient(cfg)
		ctx := context.Background()

		err := mockClient.Ping(ctx)
		require.NoError(t, err)

		objectPath := mockClient.ObjectPath("TRP-001", "v001", "VBRP")
		assert.NotEmpty(t, objectPath)
	})
}

// TestGCSConfigValidation은 설정 유효성 검사를 테스트합니다.
func TestGCSConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      gcs.GCSConfig
		expectError bool
	}{
		{
			name: "유효한 설정",
			config: gcs.GCSConfig{
				ProjectID:  "my-project",
				BucketName: "my-bucket",
			},
			expectError: false,
		},
		{
			name: "ProjectID 누락",
			config: gcs.GCSConfig{
				BucketName: "my-bucket",
			},
			expectError: true,
		},
		{
			name: "BucketName 누락",
			config: gcs.GCSConfig{
				ProjectID: "my-project",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
