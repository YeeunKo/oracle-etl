// Package middleware는 HTTP 요청 처리를 위한 미들웨어를 제공합니다
package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"oracle-etl/internal/config"
)

// 인증 제외 경로 목록
var skipAuthPaths = []string{
	"/api/health",
	"/health",
}

// NewAuthMiddleware는 API Key 및 Bearer Token 인증 미들웨어를 생성합니다
func NewAuthMiddleware(cfg *config.AuthConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 인증이 비활성화된 경우 통과
		if !cfg.Enabled {
			return c.Next()
		}

		// 인증 제외 경로 확인
		path := c.Path()
		for _, skipPath := range skipAuthPaths {
			if path == skipPath {
				return c.Next()
			}
		}

		// API Key 확인
		apiKey := c.Get("X-API-Key")
		if apiKey != "" {
			if validateAPIKey(apiKey, cfg.APIKeys) {
				c.Locals("auth_method", "api_key")
				return c.Next()
			}
			return authError(c, "유효하지 않은 API Key")
		}

		// Bearer Token 확인
		authHeader := c.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := validateBearerToken(tokenString, cfg.BearerSecret)
			if err != nil {
				return authError(c, "유효하지 않은 Bearer Token: "+err.Error())
			}

			// 사용자 정보를 컨텍스트에 저장
			if sub, ok := claims["sub"].(string); ok {
				c.Locals("user_id", sub)
			}
			c.Locals("auth_method", "bearer")
			return c.Next()
		}

		// 인증 정보 없음
		return authError(c, "인증 정보가 필요합니다")
	}
}

// validateAPIKey는 제공된 API Key가 유효한지 확인합니다
func validateAPIKey(key string, validKeys []string) bool {
	for _, validKey := range validKeys {
		if key == validKey {
			return true
		}
	}
	return false
}

// validateBearerToken는 JWT 토큰을 검증하고 클레임을 반환합니다
func validateBearerToken(tokenString, secret string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// HMAC 서명 방식 확인
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, jwt.ErrTokenMalformed
	}

	return claims, nil
}

// authError는 인증 실패 응답을 반환합니다
func authError(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"code":    "AUTHENTICATION_ERROR",
		"message": message,
	})
}
