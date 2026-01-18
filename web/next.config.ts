/**
 * Next.js 설정
 * API 프록시 및 빌드 최적화 설정
 */

import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  /**
   * API 프록시 설정
   * 개발 환경에서 Go 백엔드(:8080)로 API 요청 프록시
   */
  async rewrites() {
    return [
      {
        source: "/api/:path*",
        destination: "http://localhost:8080/api/:path*",
      },
    ];
  },

  /**
   * 실험적 기능 설정
   * typedRoutes는 모든 라우트가 정의된 후 활성화
   */
  // experimental: {
  //   typedRoutes: true,
  // },

  /**
   * 이미지 최적화 설정
   */
  images: {
    // 외부 이미지 도메인 (필요 시 추가)
    remotePatterns: [],
  },

  /**
   * 환경 변수
   */
  env: {
    // API 기본 URL (클라이언트 사이드)
    NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080",
  },

  /**
   * 헤더 설정
   */
  async headers() {
    return [
      {
        // 모든 API 경로에 CORS 헤더 추가 (개발 환경)
        source: "/api/:path*",
        headers: [
          { key: "Access-Control-Allow-Credentials", value: "true" },
          { key: "Access-Control-Allow-Origin", value: "*" },
          { key: "Access-Control-Allow-Methods", value: "GET,DELETE,PATCH,POST,PUT,OPTIONS" },
          { key: "Access-Control-Allow-Headers", value: "X-API-Key, X-CSRF-Token, X-Requested-With, Accept, Accept-Version, Content-Length, Content-MD5, Content-Type, Date, X-Api-Version" },
        ],
      },
    ];
  },
};

export default nextConfig;
