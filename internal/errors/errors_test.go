// Package errors는 구조화된 에러 처리를 제공합니다.
// M7-01: 표준 에러 응답 포맷 테스트
package errors

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewError는 새 에러 생성을 테스트합니다
func TestNewError(t *testing.T) {
	err := NewError(ErrCodeOracleConnection, "Oracle 연결 실패")

	assert.Equal(t, ErrCodeOracleConnection, err.Code)
	assert.Equal(t, "Oracle 연결 실패", err.Message)
	assert.Empty(t, err.Details)
	assert.Empty(t, err.TraceID)
}

// TestErrorResponse_WithDetails는 상세 정보 추가를 테스트합니다
func TestErrorResponse_WithDetails(t *testing.T) {
	details := map[string]any{
		"table":  "VBRP",
		"column": "MANDT",
	}

	err := NewError(ErrCodeValidation, "유효성 검사 실패").
		WithDetails(details)

	assert.Equal(t, details, err.Details)
}

// TestErrorResponse_WithTraceID는 TraceID 추가를 테스트합니다
func TestErrorResponse_WithTraceID(t *testing.T) {
	traceID := "abc-123-def-456"

	err := NewError(ErrCodeInternal, "내부 오류").
		WithTraceID(traceID)

	assert.Equal(t, traceID, err.TraceID)
}

// TestErrorResponse_Chaining은 메서드 체이닝을 테스트합니다
func TestErrorResponse_Chaining(t *testing.T) {
	err := NewError(ErrCodeGCSUpload, "업로드 실패").
		WithDetails(map[string]any{"bucket": "test-bucket"}).
		WithTraceID("trace-123")

	assert.Equal(t, ErrCodeGCSUpload, err.Code)
	assert.Equal(t, "업로드 실패", err.Message)
	assert.Equal(t, "test-bucket", err.Details.(map[string]any)["bucket"])
	assert.Equal(t, "trace-123", err.TraceID)
}

// TestErrorResponse_JSON은 JSON 직렬화를 테스트합니다
func TestErrorResponse_JSON(t *testing.T) {
	err := NewError(ErrCodeTransportNotFound, "Transport를 찾을 수 없음").
		WithTraceID("trace-xyz")

	data, jsonErr := json.Marshal(err)
	require.NoError(t, jsonErr)

	var parsed ErrorResponse
	jsonErr = json.Unmarshal(data, &parsed)
	require.NoError(t, jsonErr)

	assert.Equal(t, "TRANSPORT_NOT_FOUND", parsed.Code)
	assert.Equal(t, "Transport를 찾을 수 없음", parsed.Message)
	assert.Equal(t, "trace-xyz", parsed.TraceID)
}

// TestErrorResponse_JSONOmitsEmptyDetails는 빈 details가 JSON에서 제외되는지 테스트합니다
func TestErrorResponse_JSONOmitsEmptyDetails(t *testing.T) {
	err := NewError(ErrCodeAuth, "인증 실패")

	data, jsonErr := json.Marshal(err)
	require.NoError(t, jsonErr)

	// details 필드가 포함되지 않아야 함
	assert.NotContains(t, string(data), "details")
}

// TestErrorResponse_Error는 error 인터페이스 구현을 테스트합니다
func TestErrorResponse_Error(t *testing.T) {
	err := NewError(ErrCodeInternal, "내부 서버 오류")

	assert.Equal(t, "[INTERNAL_ERROR] 내부 서버 오류", err.Error())
}

// TestErrorResponse_Is는 errors.Is 지원을 테스트합니다
func TestErrorResponse_Is(t *testing.T) {
	err := NewError(ErrCodeOracleConnection, "연결 실패")
	target := &ErrorResponse{Code: ErrCodeOracleConnection}

	assert.True(t, errors.Is(err, target))
}

// TestErrorResponse_IsNot은 다른 코드의 에러와 매칭되지 않는지 테스트합니다
func TestErrorResponse_IsNot(t *testing.T) {
	err := NewError(ErrCodeOracleConnection, "연결 실패")
	target := &ErrorResponse{Code: ErrCodeGCSUpload}

	assert.False(t, errors.Is(err, target))
}

// TestErrorCodes는 모든 에러 코드가 정의되어 있는지 테스트합니다
func TestErrorCodes(t *testing.T) {
	codes := []string{
		ErrCodeOracleConnection,
		ErrCodeGCSUpload,
		ErrCodeTransportNotFound,
		ErrCodeValidation,
		ErrCodeAuth,
		ErrCodeRateLimit,
		ErrCodeInternal,
	}

	for _, code := range codes {
		assert.NotEmpty(t, code, "에러 코드가 정의되어야 함")
	}
}

// TestFromError는 표준 error에서 ErrorResponse 생성을 테스트합니다
func TestFromError(t *testing.T) {
	stdErr := errors.New("standard error")

	err := FromError(ErrCodeInternal, stdErr)

	assert.Equal(t, ErrCodeInternal, err.Code)
	assert.Equal(t, "standard error", err.Message)
}

// TestFromErrorWithNil은 nil 에러 처리를 테스트합니다
func TestFromErrorWithNil(t *testing.T) {
	err := FromError(ErrCodeInternal, nil)

	assert.Nil(t, err)
}

// TestNewValidationError는 유효성 검사 에러 헬퍼를 테스트합니다
func TestNewValidationError(t *testing.T) {
	err := NewValidationError("field", "필드가 비어있습니다")

	assert.Equal(t, ErrCodeValidation, err.Code)
	assert.Contains(t, err.Message, "필드가 비어있습니다")
	assert.Equal(t, "field", err.Details.(map[string]any)["field"])
}

// TestNewOracleConnectionError는 Oracle 연결 에러 헬퍼를 테스트합니다
func TestNewOracleConnectionError(t *testing.T) {
	cause := errors.New("ORA-12154: TNS could not resolve")

	err := NewOracleConnectionError(cause)

	assert.Equal(t, ErrCodeOracleConnection, err.Code)
	assert.Contains(t, err.Message, "Oracle")
}

// TestNewGCSUploadError는 GCS 업로드 에러 헬퍼를 테스트합니다
func TestNewGCSUploadError(t *testing.T) {
	cause := errors.New("network timeout")

	err := NewGCSUploadError("bucket/path/file.jsonl.gz", cause)

	assert.Equal(t, ErrCodeGCSUpload, err.Code)
	assert.Contains(t, err.Details.(map[string]any)["path"], "bucket/path/file.jsonl.gz")
}

// TestNewNotFoundError는 리소스 미발견 에러 헬퍼를 테스트합니다
func TestNewNotFoundError(t *testing.T) {
	err := NewNotFoundError("Transport", "TRP-001")

	assert.Equal(t, ErrCodeTransportNotFound, err.Code)
	assert.Contains(t, err.Message, "Transport")
	assert.Contains(t, err.Message, "TRP-001")
}

// TestHTTPStatus는 에러 코드에 따른 HTTP 상태 코드 반환을 테스트합니다
func TestHTTPStatus(t *testing.T) {
	tests := []struct {
		code     string
		expected int
	}{
		{ErrCodeValidation, 400},
		{ErrCodeAuth, 401},
		{ErrCodeRateLimit, 429},
		{ErrCodeTransportNotFound, 404},
		{ErrCodeOracleConnection, 503},
		{ErrCodeGCSUpload, 502},
		{ErrCodeInternal, 500},
	}

	for _, tc := range tests {
		t.Run(tc.code, func(t *testing.T) {
			err := NewError(tc.code, "test")
			assert.Equal(t, tc.expected, err.HTTPStatus())
		})
	}
}
