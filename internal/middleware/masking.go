package middleware

import (
	"regexp"
	"strings"
)

// 민감 데이터 패턴 정규식
var (
	// 연결 문자열에서 비밀번호 (protocol://user:password@host)
	// [^:]*를 사용하여 빈 사용자 이름도 허용 (예: redis://:password@host)
	connectionPasswordPattern = regexp.MustCompile(`(://[^:]*:)([^@]+)(@)`)

	// key=value 형식 패턴들
	passwordPattern      = regexp.MustCompile(`(?i)(password|passwd|pwd)\s*=\s*[^\s&]+`)
	apiKeyPattern        = regexp.MustCompile(`(?i)(api[_-]?key|x-api-key)\s*=\s*[^\s&]+`)
	tokenPattern         = regexp.MustCompile(`(?i)(access[_-]?token|refresh[_-]?token|token)\s*=\s*[^\s&]+`)
	secretPattern        = regexp.MustCompile(`(?i)(secret|secret[_-]?key|client[_-]?secret)\s*=\s*[^\s&]+`)
	authorizationPattern = regexp.MustCompile(`(?i)(authorization)\s*=\s*[^\s&]+`)

	// credentials 패턴 - 값에 공백이 있을 수 있음 (예: -----BEGIN RSA PRIVATE KEY-----)
	// 줄 끝까지 또는 다음 key= 패턴 전까지 캡처
	credentialsPattern = regexp.MustCompile(`(?i)(credentials|private[_-]?key)\s*=\s*(.+?)(?:\s+\w+=|$)`)

	// Bearer 토큰 패턴
	bearerPattern = regexp.MustCompile(`(?i)(bearer)\s+[^\s]+`)

	// Authorization 헤더 패턴 (Bearer 외의 다른 인증 방식)
	authHeaderPattern = regexp.MustCompile(`(?i)(Authorization:)\s+(Basic|Digest)\s+[^\s]+`)
)

// MaskSensitiveData는 입력 문자열에서 민감한 정보를 마스킹합니다
// 비밀번호, API 키, 토큰, 자격증명 등을 "****"로 대체합니다
func MaskSensitiveData(input string) string {
	if input == "" {
		return input
	}

	result := input

	// 연결 문자열 비밀번호 마스킹 (protocol://user:password@host)
	result = connectionPasswordPattern.ReplaceAllString(result, "${1}****${3}")

	// Bearer 토큰 마스킹
	result = bearerPattern.ReplaceAllStringFunc(result, func(match string) string {
		parts := strings.SplitN(match, " ", 2)
		if len(parts) == 2 {
			return parts[0] + " ****"
		}
		return match
	})

	// Authorization 헤더 마스킹 (Basic, Digest 등)
	result = authHeaderPattern.ReplaceAllStringFunc(result, func(match string) string {
		// "Authorization: Basic xxx" -> "Authorization: ****"
		if idx := strings.Index(strings.ToLower(match), "authorization:"); idx != -1 {
			return match[:idx+14] + " ****"
		}
		return "Authorization: ****"
	})

	// credentials 패턴 마스킹 (값에 공백이 있을 수 있음)
	result = credentialsPattern.ReplaceAllStringFunc(result, func(match string) string {
		eqIdx := strings.Index(match, "=")
		if eqIdx == -1 {
			return match
		}
		// 끝에 다음 key= 패턴이 있는지 확인
		afterEq := match[eqIdx+1:]
		// 공백과 key= 패턴이 있으면 그 부분은 유지
		if spaceIdx := strings.LastIndex(afterEq, " "); spaceIdx != -1 {
			remaining := strings.TrimSpace(afterEq[spaceIdx:])
			if strings.Contains(remaining, "=") {
				return match[:eqIdx+1] + "****" + afterEq[spaceIdx:]
			}
		}
		return match[:eqIdx+1] + "****"
	})

	// key=value 패턴 마스킹
	result = maskKeyValuePattern(result, passwordPattern)
	result = maskKeyValuePattern(result, apiKeyPattern)
	result = maskKeyValuePattern(result, tokenPattern)
	result = maskKeyValuePattern(result, secretPattern)
	result = maskKeyValuePattern(result, authorizationPattern)

	return result
}

// maskKeyValuePattern는 key=value 패턴을 마스킹합니다
func maskKeyValuePattern(input string, pattern *regexp.Regexp) string {
	return pattern.ReplaceAllStringFunc(input, func(match string) string {
		// "key=value" -> "key=****"
		eqIdx := strings.Index(match, "=")
		if eqIdx == -1 {
			return match
		}
		return match[:eqIdx+1] + "****"
	})
}

// MaskConnectionString는 데이터베이스 연결 문자열에서 비밀번호를 마스킹합니다
func MaskConnectionString(connStr string) string {
	return connectionPasswordPattern.ReplaceAllString(connStr, "${1}****${3}")
}
