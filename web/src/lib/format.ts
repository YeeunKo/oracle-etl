/**
 * 포맷팅 유틸리티 함수
 * 바이트, 시간, 숫자 등의 포맷팅
 */

// ============================================================================
// 바이트 포맷팅
// ============================================================================

/**
 * 바이트 단위 배열
 */
const BYTE_UNITS = ['B', 'KB', 'MB', 'GB', 'TB', 'PB'] as const;

/**
 * 바이트 크기를 사람이 읽기 쉬운 형태로 변환
 *
 * @param bytes - 바이트 수
 * @param decimals - 소수점 자릿수 (기본값: 2)
 * @returns 포맷된 문자열 (예: "1.5 GB")
 *
 * @example
 * formatBytes(1024) // "1 KB"
 * formatBytes(1536) // "1.5 KB"
 * formatBytes(1073741824) // "1 GB"
 */
export function formatBytes(bytes: number, decimals: number = 2): string {
  if (bytes === 0) return '0 B';
  if (bytes < 0) return '-' + formatBytes(-bytes, decimals);

  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;

  const i = Math.floor(Math.log(bytes) / Math.log(k));
  const index = Math.min(i, BYTE_UNITS.length - 1);

  const value = bytes / Math.pow(k, index);

  return `${value.toFixed(dm)} ${BYTE_UNITS[index]}`;
}

/**
 * 바이트 크기를 간략한 형태로 변환 (소수점 없음)
 *
 * @param bytes - 바이트 수
 * @returns 포맷된 문자열 (예: "1GB")
 */
export function formatBytesCompact(bytes: number): string {
  return formatBytes(bytes, 0).replace(' ', '');
}

// ============================================================================
// 시간 포맷팅
// ============================================================================

/**
 * 밀리초를 사람이 읽기 쉬운 시간 형태로 변환
 *
 * @param ms - 밀리초
 * @param showMs - 밀리초 표시 여부 (기본값: false)
 * @returns 포맷된 문자열 (예: "1시간 30분 45초")
 *
 * @example
 * formatDuration(5400000) // "1시간 30분"
 * formatDuration(90000) // "1분 30초"
 * formatDuration(5000) // "5초"
 */
export function formatDuration(ms: number, showMs: boolean = false): string {
  if (ms < 0) return '-' + formatDuration(-ms, showMs);
  if (ms === 0) return '0초';

  const seconds = Math.floor(ms / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);
  const days = Math.floor(hours / 24);

  const parts: string[] = [];

  if (days > 0) {
    parts.push(`${days}일`);
  }

  if (hours % 24 > 0) {
    parts.push(`${hours % 24}시간`);
  }

  if (minutes % 60 > 0) {
    parts.push(`${minutes % 60}분`);
  }

  if (seconds % 60 > 0 || parts.length === 0) {
    parts.push(`${seconds % 60}초`);
  }

  if (showMs && ms % 1000 > 0) {
    parts.push(`${ms % 1000}ms`);
  }

  return parts.join(' ');
}

/**
 * 밀리초를 간략한 시간 형태로 변환
 *
 * @param ms - 밀리초
 * @returns 포맷된 문자열 (예: "1h 30m")
 */
export function formatDurationCompact(ms: number): string {
  if (ms < 0) return '-' + formatDurationCompact(-ms);
  if (ms === 0) return '0s';

  const seconds = Math.floor(ms / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);
  const days = Math.floor(hours / 24);

  if (days > 0) {
    return `${days}d ${hours % 24}h`;
  }

  if (hours > 0) {
    return `${hours}h ${minutes % 60}m`;
  }

  if (minutes > 0) {
    return `${minutes}m ${seconds % 60}s`;
  }

  return `${seconds}s`;
}

/**
 * 초 단위를 HH:MM:SS 형태로 변환
 *
 * @param seconds - 초
 * @returns 포맷된 문자열 (예: "01:30:45")
 */
export function formatTime(seconds: number): string {
  if (seconds < 0) return '-' + formatTime(-seconds);

  const hrs = Math.floor(seconds / 3600);
  const mins = Math.floor((seconds % 3600) / 60);
  const secs = Math.floor(seconds % 60);

  if (hrs > 0) {
    return `${hrs.toString().padStart(2, '0')}:${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
  }

  return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
}

// ============================================================================
// 날짜 포맷팅
// ============================================================================

/**
 * ISO 날짜 문자열을 한국어 형식으로 변환
 *
 * @param isoString - ISO 8601 형식 날짜 문자열
 * @returns 포맷된 문자열 (예: "2025년 1월 18일 14:30")
 */
export function formatDate(isoString: string): string {
  const date = new Date(isoString);

  if (isNaN(date.getTime())) {
    return '유효하지 않은 날짜';
  }

  return date.toLocaleDateString('ko-KR', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

/**
 * ISO 날짜 문자열을 간략한 형식으로 변환
 *
 * @param isoString - ISO 8601 형식 날짜 문자열
 * @returns 포맷된 문자열 (예: "2025-01-18 14:30")
 */
export function formatDateCompact(isoString: string): string {
  const date = new Date(isoString);

  if (isNaN(date.getTime())) {
    return '유효하지 않은 날짜';
  }

  const year = date.getFullYear();
  const month = (date.getMonth() + 1).toString().padStart(2, '0');
  const day = date.getDate().toString().padStart(2, '0');
  const hours = date.getHours().toString().padStart(2, '0');
  const minutes = date.getMinutes().toString().padStart(2, '0');

  return `${year}-${month}-${day} ${hours}:${minutes}`;
}

/**
 * 상대적 시간 표시 (예: "3분 전", "2시간 전")
 *
 * @param isoString - ISO 8601 형식 날짜 문자열
 * @returns 상대적 시간 문자열
 */
export function formatRelativeTime(isoString: string): string {
  const date = new Date(isoString);

  if (isNaN(date.getTime())) {
    return '유효하지 않은 날짜';
  }

  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffSeconds = Math.floor(diffMs / 1000);
  const diffMinutes = Math.floor(diffSeconds / 60);
  const diffHours = Math.floor(diffMinutes / 60);
  const diffDays = Math.floor(diffHours / 24);

  if (diffSeconds < 0) {
    return '방금';
  }

  if (diffSeconds < 60) {
    return '방금 전';
  }

  if (diffMinutes < 60) {
    return `${diffMinutes}분 전`;
  }

  if (diffHours < 24) {
    return `${diffHours}시간 전`;
  }

  if (diffDays < 7) {
    return `${diffDays}일 전`;
  }

  if (diffDays < 30) {
    return `${Math.floor(diffDays / 7)}주 전`;
  }

  if (diffDays < 365) {
    return `${Math.floor(diffDays / 30)}개월 전`;
  }

  return `${Math.floor(diffDays / 365)}년 전`;
}

// ============================================================================
// 숫자 포맷팅
// ============================================================================

/**
 * 숫자를 천 단위 콤마가 포함된 형태로 변환
 *
 * @param value - 숫자
 * @returns 포맷된 문자열 (예: "1,234,567")
 */
export function formatNumber(value: number): string {
  return value.toLocaleString('ko-KR');
}

/**
 * 숫자를 축약된 형태로 변환 (K, M, B 단위)
 *
 * @param value - 숫자
 * @param decimals - 소수점 자릿수 (기본값: 1)
 * @returns 포맷된 문자열 (예: "1.5M")
 *
 * @example
 * formatNumberCompact(1500) // "1.5K"
 * formatNumberCompact(1500000) // "1.5M"
 * formatNumberCompact(1500000000) // "1.5B"
 */
export function formatNumberCompact(value: number, decimals: number = 1): string {
  if (value < 0) return '-' + formatNumberCompact(-value, decimals);

  const units = [
    { threshold: 1e9, suffix: 'B' },
    { threshold: 1e6, suffix: 'M' },
    { threshold: 1e3, suffix: 'K' },
  ];

  for (const { threshold, suffix } of units) {
    if (value >= threshold) {
      const formatted = (value / threshold).toFixed(decimals);
      // 불필요한 0 제거 (예: "1.0K" -> "1K")
      const cleaned = formatted.replace(/\.0+$/, '');
      return `${cleaned}${suffix}`;
    }
  }

  return value.toString();
}

/**
 * 퍼센트 형식으로 변환
 *
 * @param value - 비율 (0~1) 또는 퍼센트 (0~100)
 * @param isRatio - true면 0~1 비율, false면 0~100 퍼센트 (기본값: true)
 * @param decimals - 소수점 자릿수 (기본값: 1)
 * @returns 포맷된 문자열 (예: "75.5%")
 */
export function formatPercent(
  value: number,
  isRatio: boolean = true,
  decimals: number = 1
): string {
  const percent = isRatio ? value * 100 : value;
  return `${percent.toFixed(decimals)}%`;
}

// ============================================================================
// 행 수 포맷팅 (ETL 특화)
// ============================================================================

/**
 * 행 수를 사람이 읽기 쉬운 형태로 변환
 *
 * @param rows - 행 수
 * @returns 포맷된 문자열 (예: "1.5M rows")
 */
export function formatRows(rows: number): string {
  if (rows === 0) return '0 rows';
  if (rows === 1) return '1 row';

  const compact = formatNumberCompact(rows);
  return `${compact} rows`;
}

/**
 * 처리량(throughput)을 포맷팅
 *
 * @param rowsPerSecond - 초당 처리 행 수
 * @returns 포맷된 문자열 (예: "1.5K rows/sec")
 */
export function formatThroughput(rowsPerSecond: number): string {
  const compact = formatNumberCompact(rowsPerSecond);
  return `${compact} rows/sec`;
}
