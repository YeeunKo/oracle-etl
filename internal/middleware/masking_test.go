package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMaskSensitiveData_Password는 비밀번호 마스킹을 테스트
func TestMaskSensitiveData_Password(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "password in URL",
			input:    "oracle://user:secretpass@host:1521/db",
			expected: "oracle://user:****@host:1521/db",
		},
		{
			name:     "password= format",
			input:    "password=mysecret123",
			expected: "password=****",
		},
		{
			name:     "PASSWORD= uppercase",
			input:    "PASSWORD=MySecret",
			expected: "PASSWORD=****",
		},
		{
			name:     "pwd= format",
			input:    "pwd=secret",
			expected: "pwd=****",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskSensitiveData(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMaskSensitiveData_APIKey는 API Key 마스킹을 테스트
func TestMaskSensitiveData_APIKey(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "api_key= format",
			input:    "api_key=sk-1234567890abcdef",
			expected: "api_key=****",
		},
		{
			name:     "apikey= format",
			input:    "apikey=my-secret-key",
			expected: "apikey=****",
		},
		{
			name:     "API-Key= format",
			input:    "API-Key=secret123",
			expected: "API-Key=****",
		},
		{
			name:     "x-api-key= format",
			input:    "x-api-key=abcd1234",
			expected: "x-api-key=****",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskSensitiveData(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMaskSensitiveData_BearerToken는 Bearer 토큰 마스킹을 테스트
func TestMaskSensitiveData_BearerToken(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Bearer token",
			input:    "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.xxx",
			expected: "Authorization: Bearer ****",
		},
		{
			name:     "bearer lowercase",
			input:    "bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			expected: "bearer ****",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskSensitiveData(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMaskSensitiveData_Authorization는 Authorization 헤더 마스킹을 테스트
func TestMaskSensitiveData_Authorization(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Authorization header",
			input:    "Authorization: Basic dXNlcjpwYXNz",
			expected: "Authorization: ****",
		},
		{
			name:     "authorization= format",
			input:    "authorization=secret-token",
			expected: "authorization=****",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskSensitiveData(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMaskSensitiveData_Secret는 시크릿 키 마스킹을 테스트
func TestMaskSensitiveData_Secret(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "secret= format",
			input:    "secret=my-jwt-secret",
			expected: "secret=****",
		},
		{
			name:     "secret_key= format",
			input:    "secret_key=abc123xyz",
			expected: "secret_key=****",
		},
		{
			name:     "client_secret= format",
			input:    "client_secret=oauth-secret",
			expected: "client_secret=****",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskSensitiveData(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMaskSensitiveData_Token는 토큰 마스킹을 테스트
func TestMaskSensitiveData_Token(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "token= format",
			input:    "token=abc123def456",
			expected: "token=****",
		},
		{
			name:     "access_token= format",
			input:    "access_token=eyJhbGciOiJIUzI1NiJ9",
			expected: "access_token=****",
		},
		{
			name:     "refresh_token= format",
			input:    "refresh_token=refresh123",
			expected: "refresh_token=****",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskSensitiveData(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMaskSensitiveData_Credentials는 자격 증명 마스킹을 테스트
func TestMaskSensitiveData_Credentials(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "credentials= format",
			input:    "credentials=base64encoded",
			expected: "credentials=****",
		},
		{
			name:     "private_key= format",
			input:    "private_key=-----BEGIN RSA PRIVATE KEY-----",
			expected: "private_key=****",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskSensitiveData(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMaskSensitiveData_MultiplePatterns는 여러 패턴 동시 마스킹을 테스트
func TestMaskSensitiveData_MultiplePatterns(t *testing.T) {
	input := "connection: password=secret123 api_key=abc123 token=xyz"
	expected := "connection: password=**** api_key=**** token=****"

	result := MaskSensitiveData(input)
	assert.Equal(t, expected, result)
}

// TestMaskSensitiveData_NoSensitiveData는 민감 데이터 없는 경우를 테스트
func TestMaskSensitiveData_NoSensitiveData(t *testing.T) {
	input := "normal log message without sensitive data"
	result := MaskSensitiveData(input)
	assert.Equal(t, input, result)
}

// TestMaskSensitiveData_EmptyString는 빈 문자열 처리를 테스트
func TestMaskSensitiveData_EmptyString(t *testing.T) {
	result := MaskSensitiveData("")
	assert.Equal(t, "", result)
}

// TestMaskSensitiveData_ConnectionString는 연결 문자열 마스킹을 테스트
func TestMaskSensitiveData_ConnectionString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "PostgreSQL connection",
			input:    "postgres://user:password123@localhost:5432/db",
			expected: "postgres://user:****@localhost:5432/db",
		},
		{
			name:     "MongoDB connection",
			input:    "mongodb://admin:secretpass@mongo.example.com:27017",
			expected: "mongodb://admin:****@mongo.example.com:27017",
		},
		{
			name:     "Redis connection",
			input:    "redis://:mypassword@redis.example.com:6379",
			expected: "redis://:****@redis.example.com:6379",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskSensitiveData(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMaskConnectionString는 연결 문자열 마스킹 함수를 테스트
func TestMaskConnectionString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Oracle connection",
			input:    "oracle://admin:password123@dbhost:1521/orcl",
			expected: "oracle://admin:****@dbhost:1521/orcl",
		},
		{
			name:     "MySQL connection",
			input:    "mysql://root:secretpass@localhost:3306/mydb",
			expected: "mysql://root:****@localhost:3306/mydb",
		},
		{
			name:     "No password in URL",
			input:    "https://example.com/path",
			expected: "https://example.com/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskConnectionString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
