// Package gcs는 Google Cloud Storage 클라이언트 및 업로드 기능을 제공합니다.
package gcs

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGCSConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      GCSConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "유효한 설정",
			config: GCSConfig{
				ProjectID:       "my-project",
				BucketName:      "my-bucket",
				CredentialsFile: "/path/to/credentials.json",
				ChunkSize:       16 * 1024 * 1024, // 16MB
				Timeout:         5 * time.Minute,
			},
			expectError: false,
		},
		{
			name: "프로젝트 ID 누락",
			config: GCSConfig{
				BucketName:      "my-bucket",
				CredentialsFile: "/path/to/credentials.json",
			},
			expectError: true,
			errorMsg:    "ProjectID",
		},
		{
			name: "버킷 이름 누락",
			config: GCSConfig{
				ProjectID:       "my-project",
				CredentialsFile: "/path/to/credentials.json",
			},
			expectError: true,
			errorMsg:    "BucketName",
		},
		{
			name: "기본값 적용",
			config: GCSConfig{
				ProjectID:  "my-project",
				BucketName: "my-bucket",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGCSConfig_ApplyDefaults(t *testing.T) {
	config := GCSConfig{
		ProjectID:  "my-project",
		BucketName: "my-bucket",
	}

	config.ApplyDefaults()

	assert.Equal(t, DefaultChunkSize, config.ChunkSize)
	assert.Equal(t, DefaultTimeout, config.Timeout)
}

func TestNewMockClient(t *testing.T) {
	config := GCSConfig{
		ProjectID:  "test-project",
		BucketName: "test-bucket",
	}

	client := NewMockClient(config)
	require.NotNil(t, client)

	// Mock 클라이언트는 연결 테스트가 성공해야 함
	ctx := context.Background()
	err := client.Ping(ctx)
	require.NoError(t, err)
}

func TestMockClient_ObjectPath(t *testing.T) {
	config := GCSConfig{
		ProjectID:  "test-project",
		BucketName: "test-bucket",
	}

	client := NewMockClient(config)

	tests := []struct {
		transportID string
		jobVersion  string
		tableName   string
		expected    string
	}{
		{
			transportID: "TRP-001",
			jobVersion:  "v001",
			tableName:   "VBRP",
			expected:    "TRP-001/v001/VBRP.jsonl.gz",
		},
		{
			transportID: "TRP-20260118-001",
			jobVersion:  "v002",
			tableName:   "LIKP",
			expected:    "TRP-20260118-001/v002/LIKP.jsonl.gz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.tableName, func(t *testing.T) {
			path := client.ObjectPath(tt.transportID, tt.jobVersion, tt.tableName)
			assert.Equal(t, tt.expected, path)
		})
	}
}

func TestMockClient_FullGCSPath(t *testing.T) {
	config := GCSConfig{
		ProjectID:  "test-project",
		BucketName: "oracle-etl-data",
	}

	client := NewMockClient(config)

	path := client.FullGCSPath("TRP-001", "v001", "VBRP")
	expected := "gs://oracle-etl-data/TRP-001/v001/VBRP.jsonl.gz"
	assert.Equal(t, expected, path)
}

func TestMockClient_BucketName(t *testing.T) {
	config := GCSConfig{
		ProjectID:  "test-project",
		BucketName: "my-test-bucket",
	}

	client := NewMockClient(config)
	assert.Equal(t, "my-test-bucket", client.BucketName())
}

func TestMockClient_Close(t *testing.T) {
	config := GCSConfig{
		ProjectID:  "test-project",
		BucketName: "test-bucket",
	}

	client := NewMockClient(config)

	err := client.Close()
	require.NoError(t, err)
}
