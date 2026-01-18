package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoadConfig_FromFile은 config.yaml 파일에서 설정을 로드하는 테스트
func TestLoadConfig_FromFile(t *testing.T) {
	// 임시 config 파일 생성
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
server:
  port: 8080
  read_timeout: 30s
  write_timeout: 300s
app:
  name: oracle-etl
  version: 1.0.0
  environment: development
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// 설정 로드
	cfg, err := Load(configPath)
	require.NoError(t, err)

	// 서버 설정 검증
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "30s", cfg.Server.ReadTimeout)
	assert.Equal(t, "300s", cfg.Server.WriteTimeout)

	// 앱 설정 검증
	assert.Equal(t, "oracle-etl", cfg.App.Name)
	assert.Equal(t, "1.0.0", cfg.App.Version)
	assert.Equal(t, "development", cfg.App.Environment)
}

// TestLoadConfig_EnvOverride는 환경 변수가 설정을 오버라이드하는지 테스트
func TestLoadConfig_EnvOverride(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
server:
  port: 8080
app:
  version: 1.0.0
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// 환경 변수 설정
	os.Setenv("SERVER_PORT", "9090")
	defer os.Unsetenv("SERVER_PORT")

	cfg, err := Load(configPath)
	require.NoError(t, err)

	// 환경 변수로 오버라이드된 값 검증
	assert.Equal(t, 9090, cfg.Server.Port)
}

// TestLoadConfig_Defaults는 기본값이 설정되는지 테스트
func TestLoadConfig_Defaults(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// 빈 설정 파일
	configContent := ``
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	cfg, err := Load(configPath)
	require.NoError(t, err)

	// 기본값 검증 (300s로 변경 - ETL 작업을 위해)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "10s", cfg.Server.ReadTimeout)
	assert.Equal(t, "300s", cfg.Server.WriteTimeout)
	assert.Equal(t, "production", cfg.App.Environment)
}

// TestLoadConfig_OracleDefaults는 Oracle 기본 설정을 테스트합니다
func TestLoadConfig_OracleDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := ``
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	cfg, err := Load(configPath)
	require.NoError(t, err)

	// Oracle 기본값 검증
	assert.Equal(t, 2, cfg.Oracle.PoolMin)
	assert.Equal(t, 10, cfg.Oracle.PoolMax)
	assert.Equal(t, 1000, cfg.Oracle.FetchArraySize)

	// ETL 기본값 검증
	assert.Equal(t, 10000, cfg.ETL.ChunkSize)
	assert.Equal(t, 4, cfg.ETL.ParallelTables)
	assert.Equal(t, 3, cfg.ETL.RetryAttempts)
}

// TestLoadConfig_OracleFromEnv는 Oracle 환경 변수 바인딩을 테스트합니다
func TestLoadConfig_OracleFromEnv(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	err := os.WriteFile(configPath, []byte(""), 0644)
	require.NoError(t, err)

	// 환경 변수 설정
	os.Setenv("ORACLE_TNS_NAME", "test_tns")
	os.Setenv("ORACLE_USERNAME", "test_user")
	defer os.Unsetenv("ORACLE_TNS_NAME")
	defer os.Unsetenv("ORACLE_USERNAME")

	cfg, err := Load(configPath)
	require.NoError(t, err)

	assert.Equal(t, "test_tns", cfg.Oracle.TNSName)
	assert.Equal(t, "test_user", cfg.Oracle.Username)
}

// TestConfig_Validate는 설정 유효성 검사 테스트
func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "유효한 설정",
			config: Config{
				Server: ServerConfig{
					Port:         8080,
					ReadTimeout:  "10s",
					WriteTimeout: "60s",
				},
				App: AppConfig{
					Name:        "test-app",
					Version:     "1.0.0",
					Environment: "development",
				},
			},
			wantErr: false,
		},
		{
			name: "잘못된 포트 - 음수",
			config: Config{
				Server: ServerConfig{
					Port: -1,
				},
			},
			wantErr: true,
		},
		{
			name: "잘못된 포트 - 65535 초과",
			config: Config{
				Server: ServerConfig{
					Port: 70000,
				},
			},
			wantErr: true,
		},
		{
			name: "유효한 Oracle 설정",
			config: Config{
				Server: ServerConfig{
					Port: 8080,
				},
				Oracle: OracleConfig{
					TNSName:  "test_tns",
					Username: "test_user",
					PoolMin:  2,
					PoolMax:  10,
				},
			},
			wantErr: false,
		},
		{
			name: "Oracle 설정 - username 없음",
			config: Config{
				Server: ServerConfig{
					Port: 8080,
				},
				Oracle: OracleConfig{
					TNSName: "test_tns",
					// Username 없음
					PoolMin: 2,
					PoolMax: 10,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestConfig_HasOracleConfig는 Oracle 설정 존재 여부 확인 테스트
func TestConfig_HasOracleConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		expected bool
	}{
		{
			name: "Oracle 설정 있음",
			config: Config{
				Oracle: OracleConfig{
					TNSName:  "test_tns",
					Username: "test_user",
				},
			},
			expected: true,
		},
		{
			name: "Oracle 설정 없음",
			config: Config{
				Oracle: OracleConfig{},
			},
			expected: false,
		},
		{
			name: "TNS만 있음",
			config: Config{
				Oracle: OracleConfig{
					TNSName: "test_tns",
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.HasOracleConfig()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestLoadConfig_GCSDefaults는 GCS 기본 설정을 테스트합니다
func TestLoadConfig_GCSDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := ``
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	cfg, err := Load(configPath)
	require.NoError(t, err)

	// GCS 기본값 검증
	assert.Equal(t, 16*1024*1024, cfg.GCS.ChunkSize) // 16MB
	assert.Equal(t, 600, cfg.GCS.TimeoutSeconds)     // 10분
}

// TestLoadConfig_GCSFromEnv는 GCS 환경 변수 바인딩을 테스트합니다
func TestLoadConfig_GCSFromEnv(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	err := os.WriteFile(configPath, []byte(""), 0644)
	require.NoError(t, err)

	// 환경 변수 설정
	os.Setenv("GCS_PROJECT_ID", "test-project")
	os.Setenv("GCS_BUCKET_NAME", "test-bucket")
	os.Setenv("GCS_CREDENTIALS_FILE", "/path/to/credentials.json")
	defer os.Unsetenv("GCS_PROJECT_ID")
	defer os.Unsetenv("GCS_BUCKET_NAME")
	defer os.Unsetenv("GCS_CREDENTIALS_FILE")

	cfg, err := Load(configPath)
	require.NoError(t, err)

	assert.Equal(t, "test-project", cfg.GCS.ProjectID)
	assert.Equal(t, "test-bucket", cfg.GCS.BucketName)
	assert.Equal(t, "/path/to/credentials.json", cfg.GCS.CredentialsFile)
}

// TestConfig_GCSValidation은 GCS 설정 유효성 검사를 테스트합니다
func TestConfig_GCSValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "유효한 GCS 설정",
			config: Config{
				Server: ServerConfig{Port: 8080},
				GCS: GCSConfig{
					ProjectID:  "test-project",
					BucketName: "test-bucket",
				},
			},
			wantErr: false,
		},
		{
			name: "GCS 설정 없음 (유효)",
			config: Config{
				Server: ServerConfig{Port: 8080},
				GCS:    GCSConfig{},
			},
			wantErr: false,
		},
		{
			name: "BucketName만 있고 ProjectID 없음",
			config: Config{
				Server: ServerConfig{Port: 8080},
				GCS: GCSConfig{
					BucketName: "test-bucket",
					// ProjectID 없음
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestConfig_HasGCSConfig는 GCS 설정 존재 여부 확인 테스트
func TestConfig_HasGCSConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		expected bool
	}{
		{
			name: "GCS 설정 있음",
			config: Config{
				GCS: GCSConfig{
					ProjectID:  "test-project",
					BucketName: "test-bucket",
				},
			},
			expected: true,
		},
		{
			name: "GCS 설정 없음",
			config: Config{
				GCS: GCSConfig{},
			},
			expected: false,
		},
		{
			name: "ProjectID만 있음",
			config: Config{
				GCS: GCSConfig{
					ProjectID: "test-project",
				},
			},
			expected: false,
		},
		{
			name: "BucketName만 있음",
			config: Config{
				GCS: GCSConfig{
					BucketName: "test-bucket",
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.HasGCSConfig()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestConfig_GetGCSTimeout은 GCS 타임아웃 변환을 테스트합니다
func TestConfig_GetGCSTimeout(t *testing.T) {
	tests := []struct {
		name           string
		timeoutSeconds int
		expectedMins   int
	}{
		{
			name:           "기본값 (0일 때)",
			timeoutSeconds: 0,
			expectedMins:   10, // 기본값 10분
		},
		{
			name:           "음수일 때",
			timeoutSeconds: -1,
			expectedMins:   10, // 기본값 10분
		},
		{
			name:           "5분 설정",
			timeoutSeconds: 300,
			expectedMins:   5,
		},
		{
			name:           "15분 설정",
			timeoutSeconds: 900,
			expectedMins:   15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				GCS: GCSConfig{
					TimeoutSeconds: tt.timeoutSeconds,
				},
			}
			timeout := cfg.GetGCSTimeout()
			expectedDuration := time.Duration(tt.expectedMins) * time.Minute
			assert.Equal(t, expectedDuration, timeout)
		})
	}
}
