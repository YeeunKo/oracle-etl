// Package errors는 구조화된 에러 처리를 제공합니다.
// M7-01: 표준 에러 응답 포맷
package errors

import (
	"fmt"
	"net/http"
)

// 에러 코드 상수
const (
	// ErrCodeOracleConnection은 Oracle 연결 오류를 나타냅니다
	ErrCodeOracleConnection = "ORACLE_CONNECTION_ERROR"
	// ErrCodeGCSUpload은 GCS 업로드 오류를 나타냅니다
	ErrCodeGCSUpload = "GCS_UPLOAD_ERROR"
	// ErrCodeTransportNotFound는 Transport를 찾을 수 없음을 나타냅니다
	ErrCodeTransportNotFound = "TRANSPORT_NOT_FOUND"
	// ErrCodeValidation은 유효성 검사 오류를 나타냅니다
	ErrCodeValidation = "VALIDATION_ERROR"
	// ErrCodeAuth는 인증 오류를 나타냅니다
	ErrCodeAuth = "AUTHENTICATION_ERROR"
	// ErrCodeRateLimit은 요청 제한 초과를 나타냅니다
	ErrCodeRateLimit = "RATE_LIMIT_EXCEEDED"
	// ErrCodeInternal은 내부 서버 오류를 나타냅니다
	ErrCodeInternal = "INTERNAL_ERROR"
)

// ErrorResponse는 구조화된 에러 응답 구조체입니다
type ErrorResponse struct {
	Code    string `json:"code"`              // 에러 코드 (예: ORACLE_CONNECTION_ERROR)
	Message string `json:"message"`           // 사용자 친화적 메시지
	Details any    `json:"details,omitempty"` // 추가 상세 정보 (선택)
	TraceID string `json:"trace_id,omitempty"` // 추적 ID (선택)
}

// NewError는 새로운 ErrorResponse를 생성합니다
func NewError(code, message string) *ErrorResponse {
	return &ErrorResponse{
		Code:    code,
		Message: message,
	}
}

// WithDetails는 상세 정보를 추가하고 체이닝을 위해 자신을 반환합니다
func (e *ErrorResponse) WithDetails(details any) *ErrorResponse {
	e.Details = details
	return e
}

// WithTraceID는 TraceID를 추가하고 체이닝을 위해 자신을 반환합니다
func (e *ErrorResponse) WithTraceID(traceID string) *ErrorResponse {
	e.TraceID = traceID
	return e
}

// Error는 error 인터페이스를 구현합니다
func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Is는 errors.Is 지원을 위한 메서드입니다
// 에러 코드가 같으면 동일한 에러로 간주합니다
func (e *ErrorResponse) Is(target error) bool {
	t, ok := target.(*ErrorResponse)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// HTTPStatus는 에러 코드에 따른 HTTP 상태 코드를 반환합니다
func (e *ErrorResponse) HTTPStatus() int {
	switch e.Code {
	case ErrCodeValidation:
		return http.StatusBadRequest // 400
	case ErrCodeAuth:
		return http.StatusUnauthorized // 401
	case ErrCodeTransportNotFound:
		return http.StatusNotFound // 404
	case ErrCodeRateLimit:
		return http.StatusTooManyRequests // 429
	case ErrCodeOracleConnection:
		return http.StatusServiceUnavailable // 503
	case ErrCodeGCSUpload:
		return http.StatusBadGateway // 502
	case ErrCodeInternal:
		return http.StatusInternalServerError // 500
	default:
		return http.StatusInternalServerError // 500
	}
}

// FromError는 표준 error에서 ErrorResponse를 생성합니다
func FromError(code string, err error) *ErrorResponse {
	if err == nil {
		return nil
	}
	return NewError(code, err.Error())
}

// NewValidationError는 유효성 검사 에러를 생성합니다
func NewValidationError(field, message string) *ErrorResponse {
	return NewError(ErrCodeValidation, message).
		WithDetails(map[string]any{
			"field": field,
		})
}

// NewOracleConnectionError는 Oracle 연결 에러를 생성합니다
func NewOracleConnectionError(cause error) *ErrorResponse {
	msg := "Oracle 데이터베이스 연결 실패"
	if cause != nil {
		msg = fmt.Sprintf("Oracle 연결 오류: %v", cause)
	}
	return NewError(ErrCodeOracleConnection, msg).
		WithDetails(map[string]any{
			"cause": cause.Error(),
		})
}

// NewGCSUploadError는 GCS 업로드 에러를 생성합니다
func NewGCSUploadError(path string, cause error) *ErrorResponse {
	msg := "GCS 업로드 실패"
	if cause != nil {
		msg = fmt.Sprintf("GCS 업로드 오류: %v", cause)
	}
	return NewError(ErrCodeGCSUpload, msg).
		WithDetails(map[string]any{
			"path":  path,
			"cause": cause.Error(),
		})
}

// NewNotFoundError는 리소스 미발견 에러를 생성합니다
func NewNotFoundError(resourceType, id string) *ErrorResponse {
	return NewError(ErrCodeTransportNotFound,
		fmt.Sprintf("%s '%s'을(를) 찾을 수 없습니다", resourceType, id))
}

// NewAuthError는 인증 에러를 생성합니다
func NewAuthError(message string) *ErrorResponse {
	return NewError(ErrCodeAuth, message)
}

// NewRateLimitError는 요청 제한 에러를 생성합니다
func NewRateLimitError() *ErrorResponse {
	return NewError(ErrCodeRateLimit, "요청 제한 초과. 잠시 후 다시 시도해주세요.")
}

// NewInternalError는 내부 서버 에러를 생성합니다
func NewInternalError(message string) *ErrorResponse {
	return NewError(ErrCodeInternal, message)
}
