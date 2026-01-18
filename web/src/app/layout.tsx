/**
 * 루트 레이아웃
 * 전역 Provider 및 스타일 설정
 */

import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";

import { QueryProvider } from "@/components/providers/query-provider";
import { Toaster } from "sonner";

// ============================================================================
// 폰트 설정
// ============================================================================

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

// ============================================================================
// 메타데이터
// ============================================================================

export const metadata: Metadata = {
  title: {
    default: "Oracle ETL Pipeline",
    template: "%s | Oracle ETL",
  },
  description: "Oracle Database ETL Pipeline 관리 시스템",
  keywords: ["Oracle", "ETL", "Pipeline", "Data", "Transfer"],
  authors: [{ name: "Oracle ETL Team" }],
};

// ============================================================================
// 루트 레이아웃 컴포넌트
// ============================================================================

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="ko" suppressHydrationWarning>
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased`}
      >
        <QueryProvider>
          {children}
          <Toaster
            position="top-right"
            richColors
            closeButton
            duration={5000}
          />
        </QueryProvider>
      </body>
    </html>
  );
}
