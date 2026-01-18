// Package config는 애플리케이션 설정 관리를 담당합니다.
// Viper를 사용하여 config.yaml 파일과 환경 변수에서 설정을 로드합니다.
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config는 애플리케이션 전체 설정 구조체입니다
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	App       AppConfig       `mapstructure:"app"`
	Oracle    OracleConfig    `mapstructure:"oracle"`
	GCS       GCSConfig       `mapstructure:"gcs"`
	ETL       ETLConfig       `mapstructure:"etl"`
	Auth      AuthConfig      `mapstructure:"auth"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
	CORS      CORSConfig      `mapstructure:"cors"`
}

// ServerConfig는 HTTP 서버 관련 설정입니다
type ServerConfig struct {
	Port         int    `mapstructure:"port"`
	ReadTimeout  string `mapstructure:"read_timeout"`
	WriteTimeout string `mapstructure:"write_timeout"`
}

// AppConfig는 애플리케이션 메타데이터 설정입니다
type AppConfig struct {
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"`
}

// OracleConfig는 Oracle 데이터베이스 연결 설정입니다
type OracleConfig struct {
	WalletPath     string `mapstructure:"wallet_path"`      // mTLS wallet 디렉토리 경로
	TNSName        string `mapstructure:"tns_name"`         // TNS 이름 (예: "oracledb_high")
	Username       string `mapstructure:"username"`         // Oracle 사용자 이름
	Password       string `mapstructure:"password"`         // Oracle 비밀번호
	PoolMin        int    `mapstructure:"pool_min"`         // 최소 커넥션 수
	PoolMax        int    `mapstructure:"pool_max"`         // 최대 커넥션 수
	FetchArraySize int    `mapstructure:"fetch_array_size"` // 배치 페치 크기
	PrefetchCount  int    `mapstructure:"prefetch_count"`   // 프리페치 카운트
	DefaultOwner   string `mapstructure:"default_owner"`    // 기본 스키마 소유자
}

// GCSConfig는 Google Cloud Storage 연결 설정입니다
type GCSConfig struct {
	ProjectID       string `mapstructure:"project_id"`       // GCP 프로젝트 ID
	BucketName      string `mapstructure:"bucket_name"`      // 대상 버킷 이름
	CredentialsFile string `mapstructure:"credentials_file"` // 서비스 계정 JSON 파일 경로
	ChunkSize       int    `mapstructure:"chunk_size"`       // resumable 업로드 청크 크기 (바이트)
	TimeoutSeconds  int    `mapstructure:"timeout_seconds"`  // 작업 타임아웃 (초)
}

// ETLConfig는 ETL 작업 관련 설정입니다
type ETLConfig struct {
	ChunkSize      int    `mapstructure:"chunk_size"`      // 청크당 row 수
	ParallelTables int    `mapstructure:"parallel_tables"` // 병렬 처리 테이블 수
	RetryAttempts  int    `mapstructure:"retry_attempts"`  // 재시도 횟수
	RetryBackoff   string `mapstructure:"retry_backoff"`   // 재시도 간격
}

// AuthConfig는 API 인증 관련 설정입니다
type AuthConfig struct {
	Enabled      bool     `mapstructure:"enabled"`       // 인증 활성화 여부
	APIKeys      []string `mapstructure:"api_keys"`      // 유효한 API Key 목록
	BearerSecret string   `mapstructure:"bearer_secret"` // JWT 서명용 비밀키
}

// RateLimitConfig는 요청 제한 관련 설정입니다
type RateLimitConfig struct {
	Enabled           bool `mapstructure:"enabled"`             // Rate limiting 활성화 여부
	RequestsPerMinute int  `mapstructure:"requests_per_minute"` // 분당 최대 요청 수
	BurstSize         int  `mapstructure:"burst_size"`          // 버스트 허용 크기
}

// CORSConfig는 CORS 관련 설정입니다
type CORSConfig struct {
	Enabled          bool     `mapstructure:"enabled"`           // CORS 활성화 여부
	AllowOrigins     []string `mapstructure:"allow_origins"`     // 허용된 Origin 목록
	AllowMethods     []string `mapstructure:"allow_methods"`     // 허용된 HTTP 메서드 목록
	AllowHeaders     []string `mapstructure:"allow_headers"`     // 허용된 헤더 목록
	AllowCredentials bool     `mapstructure:"allow_credentials"` // 자격 증명 허용 여부
	ExposeHeaders    []string `mapstructure:"expose_headers"`    // 노출할 헤더 목록
	MaxAge           int      `mapstructure:"max_age"`           // preflight 캐시 시간 (초)
}

// Load는 지정된 경로의 설정 파일과 환경 변수에서 설정을 로드합니다
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// 기본값 설정
	setDefaults(v)

	// 설정 파일 읽기
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	// 설정 파일 읽기 시도 - 파일이 없거나 읽기 실패해도 기본값 사용
	// 빈 파일이나 파일 없음은 정상적인 상황으로 처리
	_ = v.ReadInConfig()

	// 환경 변수 바인딩
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 환경 변수 명시적 바인딩
	bindEnvVars(v)

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("설정 파싱 실패: %w", err)
	}

	return &cfg, nil
}

// bindEnvVars는 환경 변수를 설정에 바인딩합니다
func bindEnvVars(v *viper.Viper) {
	// 서버 설정
	_ = v.BindEnv("server.port", "SERVER_PORT")

	// Oracle 설정
	_ = v.BindEnv("oracle.wallet_path", "ORACLE_WALLET_PATH")
	_ = v.BindEnv("oracle.tns_name", "ORACLE_TNS_NAME")
	_ = v.BindEnv("oracle.username", "ORACLE_USERNAME")
	_ = v.BindEnv("oracle.password", "ORACLE_PASSWORD")
	_ = v.BindEnv("oracle.pool_min", "ORACLE_POOL_MIN")
	_ = v.BindEnv("oracle.pool_max", "ORACLE_POOL_MAX")
	_ = v.BindEnv("oracle.fetch_array_size", "ORACLE_FETCH_ARRAY_SIZE")
	_ = v.BindEnv("oracle.default_owner", "ORACLE_DEFAULT_OWNER")

	// GCS 설정
	_ = v.BindEnv("gcs.project_id", "GCS_PROJECT_ID")
	_ = v.BindEnv("gcs.bucket_name", "GCS_BUCKET_NAME")
	_ = v.BindEnv("gcs.credentials_file", "GCS_CREDENTIALS_FILE")
	_ = v.BindEnv("gcs.chunk_size", "GCS_CHUNK_SIZE")
	_ = v.BindEnv("gcs.timeout_seconds", "GCS_TIMEOUT_SECONDS")

	// Auth 설정
	_ = v.BindEnv("auth.enabled", "AUTH_ENABLED")
	_ = v.BindEnv("auth.api_keys", "AUTH_API_KEYS")
	_ = v.BindEnv("auth.bearer_secret", "AUTH_BEARER_SECRET")

	// Rate Limit 설정
	_ = v.BindEnv("rate_limit.enabled", "RATE_LIMIT_ENABLED")
	_ = v.BindEnv("rate_limit.requests_per_minute", "RATE_LIMIT_REQUESTS_PER_MINUTE")
	_ = v.BindEnv("rate_limit.burst_size", "RATE_LIMIT_BURST_SIZE")

	// CORS 설정
	_ = v.BindEnv("cors.enabled", "CORS_ENABLED")
	_ = v.BindEnv("cors.allow_origins", "CORS_ALLOW_ORIGINS")
}

// setDefaults는 Viper에 기본값을 설정합니다
func setDefaults(v *viper.Viper) {
	// 서버 기본값
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", "10s")
	v.SetDefault("server.write_timeout", "300s") // ETL 작업을 위해 늘림

	// 앱 기본값
	v.SetDefault("app.name", "oracle-etl")
	v.SetDefault("app.version", "1.0.0")
	v.SetDefault("app.environment", "production")

	// Oracle 기본값
	v.SetDefault("oracle.pool_min", 2)
	v.SetDefault("oracle.pool_max", 10)
	v.SetDefault("oracle.fetch_array_size", 1000)
	v.SetDefault("oracle.prefetch_count", 1000)

	// GCS 기본값
	v.SetDefault("gcs.chunk_size", 16*1024*1024) // 16MB
	v.SetDefault("gcs.timeout_seconds", 600)     // 10분

	// ETL 기본값
	v.SetDefault("etl.chunk_size", 10000)
	v.SetDefault("etl.parallel_tables", 4)
	v.SetDefault("etl.retry_attempts", 3)
	v.SetDefault("etl.retry_backoff", "1s")

	// Auth 기본값
	v.SetDefault("auth.enabled", false)
	v.SetDefault("auth.api_keys", []string{})
	v.SetDefault("auth.bearer_secret", "")

	// Rate Limit 기본값
	v.SetDefault("rate_limit.enabled", false)
	v.SetDefault("rate_limit.requests_per_minute", 100)
	v.SetDefault("rate_limit.burst_size", 20)

	// CORS 기본값
	v.SetDefault("cors.enabled", false)
	v.SetDefault("cors.allow_origins", []string{"*"})
	v.SetDefault("cors.allow_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	v.SetDefault("cors.allow_headers", []string{"Origin", "Content-Type", "Accept", "Authorization", "X-API-Key"})
	v.SetDefault("cors.allow_credentials", false)
	v.SetDefault("cors.expose_headers", []string{})
	v.SetDefault("cors.max_age", 86400) // 24시간
}

// Validate는 설정의 유효성을 검사합니다
func (c *Config) Validate() error {
	// 포트 유효성 검사
	if c.Server.Port < 0 || c.Server.Port > 65535 {
		return fmt.Errorf("잘못된 포트 번호: %d (0-65535 범위여야 함)", c.Server.Port)
	}

	// Oracle 설정 유효성 검사 (선택적 - 설정되지 않은 경우는 건너뜀)
	if c.Oracle.TNSName != "" {
		if c.Oracle.Username == "" {
			return fmt.Errorf("Oracle 사용자 이름이 설정되지 않음")
		}
		if c.Oracle.PoolMin < 0 {
			return fmt.Errorf("Oracle 최소 풀 크기는 0 이상이어야 함")
		}
		if c.Oracle.PoolMax < c.Oracle.PoolMin {
			return fmt.Errorf("Oracle 최대 풀 크기는 최소 풀 크기보다 크거나 같아야 함")
		}
	}

	// GCS 설정 유효성 검사 (선택적)
	if c.GCS.BucketName != "" {
		if c.GCS.ProjectID == "" {
			return fmt.Errorf("GCS 버킷이 설정되었지만 프로젝트 ID가 없음")
		}
	}

	// Auth 설정 유효성 검사
	if c.Auth.Enabled {
		if len(c.Auth.APIKeys) == 0 && c.Auth.BearerSecret == "" {
			return fmt.Errorf("Auth가 활성화되었지만 API 키 또는 Bearer 비밀키가 설정되지 않음")
		}
	}

	// Rate Limit 설정 유효성 검사
	if c.RateLimit.Enabled {
		if c.RateLimit.RequestsPerMinute <= 0 {
			return fmt.Errorf("Rate Limit가 활성화되었지만 requests_per_minute이 0 이하임")
		}
	}

	return nil
}

// HasOracleConfig는 Oracle 설정이 있는지 확인합니다
func (c *Config) HasOracleConfig() bool {
	return c.Oracle.TNSName != "" && c.Oracle.Username != ""
}

// HasGCSConfig는 GCS 설정이 있는지 확인합니다
func (c *Config) HasGCSConfig() bool {
	return c.GCS.ProjectID != "" && c.GCS.BucketName != ""
}

// GetGCSTimeout는 GCS 타임아웃을 time.Duration으로 반환합니다
func (c *Config) GetGCSTimeout() time.Duration {
	if c.GCS.TimeoutSeconds <= 0 {
		return 10 * time.Minute // 기본값
	}
	return time.Duration(c.GCS.TimeoutSeconds) * time.Second
}
