/**
 * 설정 페이지
 * API 키 관리 및 연결 상태 확인
 */

'use client';

import { useState, useEffect } from 'react';
import { MainLayout } from '@/components/layout/main-layout';
import { useAuthStore } from '@/store/auth';
import { getOracleStatus } from '@/lib/api';
import type { OracleStatus } from '@/types/api';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import {
  Key,
  Eye,
  EyeOff,
  Check,
  X,
  RefreshCw,
  Trash2,
  AlertCircle,
  CheckCircle2,
  Loader2,
} from 'lucide-react';
import { toast } from 'sonner';

// ============================================================================
// 설정 페이지 컴포넌트
// ============================================================================

/**
 * API 키를 마스킹하여 표시
 * @param key - API 키
 * @returns 마스킹된 API 키 (앞 4자리와 뒤 4자리만 표시)
 */
function maskApiKey(key: string): string {
  if (key.length <= 8) {
    return '*'.repeat(key.length);
  }
  const start = key.slice(0, 4);
  const end = key.slice(-4);
  const middle = '*'.repeat(Math.min(key.length - 8, 16));
  return `${start}${middle}${end}`;
}

/**
 * 설정 페이지
 * - API 키 설정 및 관리
 * - Oracle 연결 상태 테스트
 */
export default function SettingsPage() {
  // 인증 스토어
  const { apiKey, setApiKey, clearApiKey, isAuthenticated } = useAuthStore();

  // 로컬 상태
  const [inputKey, setInputKey] = useState('');
  const [showKey, setShowKey] = useState(false);
  const [isTestingConnection, setIsTestingConnection] = useState(false);
  const [connectionStatus, setConnectionStatus] = useState<OracleStatus | null>(null);
  const [connectionError, setConnectionError] = useState<string | null>(null);

  // 컴포넌트 마운트 시 연결 상태 초기화
  useEffect(() => {
    // 하이드레이션 이슈 방지
    setConnectionStatus(null);
    setConnectionError(null);
  }, []);

  /**
   * API 키 저장 핸들러
   */
  const handleSaveApiKey = async () => {
    if (!inputKey.trim()) {
      toast.error('API 키를 입력해주세요.');
      return;
    }

    setApiKey(inputKey.trim());
    setInputKey('');
    toast.success('API 키가 저장되었습니다.');

    // 저장 후 연결 테스트
    await testConnection();
  };

  /**
   * API 키 삭제 핸들러
   */
  const handleClearApiKey = () => {
    clearApiKey();
    setConnectionStatus(null);
    setConnectionError(null);
    toast.success('API 키가 삭제되었습니다.');
  };

  /**
   * 연결 상태 테스트 핸들러
   */
  const testConnection = async () => {
    setIsTestingConnection(true);
    setConnectionError(null);

    try {
      const status = await getOracleStatus();
      setConnectionStatus(status);

      if (status.connected) {
        toast.success('Oracle 데이터베이스에 연결되었습니다.');
      } else {
        toast.error(`연결 실패: ${status.error || '알 수 없는 오류'}`);
      }
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : '연결 테스트 중 오류가 발생했습니다.';
      setConnectionError(errorMessage);
      setConnectionStatus(null);
      toast.error(errorMessage);
    } finally {
      setIsTestingConnection(false);
    }
  };

  return (
    <MainLayout>
      <div className="space-y-6">
        {/* 페이지 헤더 */}
        <div>
          <h1 className="text-2xl font-bold tracking-tight">설정</h1>
          <p className="text-muted-foreground">
            API 키 및 연결 설정을 관리합니다.
          </p>
        </div>

        {/* API 키 관리 카드 */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Key className="h-5 w-5" />
              API 키 관리
            </CardTitle>
            <CardDescription>
              Oracle ETL Pipeline API에 접근하기 위한 API 키를 설정합니다.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {/* 현재 API 키 표시 */}
            {isAuthenticated && apiKey ? (
              <div className="space-y-4">
                <div className="flex items-center gap-2">
                  <Badge variant="secondary" className="gap-1">
                    <CheckCircle2 className="h-3 w-3" />
                    설정됨
                  </Badge>
                </div>

                {/* 마스킹된 API 키 표시 */}
                <div className="flex items-center gap-2 rounded-md border bg-muted/50 p-3">
                  <code className="flex-1 font-mono text-sm">
                    {showKey ? apiKey : maskApiKey(apiKey)}
                  </code>
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    onClick={() => setShowKey(!showKey)}
                    aria-label={showKey ? 'API 키 숨기기' : 'API 키 보기'}
                  >
                    {showKey ? (
                      <EyeOff className="h-4 w-4" />
                    ) : (
                      <Eye className="h-4 w-4" />
                    )}
                  </Button>
                </div>

                {/* 액션 버튼 */}
                <div className="flex gap-2">
                  <Button
                    variant="outline"
                    onClick={testConnection}
                    disabled={isTestingConnection}
                  >
                    {isTestingConnection ? (
                      <Loader2 className="h-4 w-4 animate-spin" />
                    ) : (
                      <RefreshCw className="h-4 w-4" />
                    )}
                    연결 테스트
                  </Button>
                  <Button
                    variant="destructive"
                    onClick={handleClearApiKey}
                  >
                    <Trash2 className="h-4 w-4" />
                    API 키 삭제
                  </Button>
                </div>
              </div>
            ) : (
              /* API 키 입력 폼 */
              <div className="space-y-4">
                <Alert>
                  <AlertCircle className="h-4 w-4" />
                  <AlertTitle>API 키 필요</AlertTitle>
                  <AlertDescription>
                    API 키가 설정되지 않았습니다. API 키를 입력하여 서버에 연결하세요.
                  </AlertDescription>
                </Alert>

                <div className="flex gap-2">
                  <Input
                    type="password"
                    placeholder="API 키를 입력하세요"
                    value={inputKey}
                    onChange={(e) => setInputKey(e.target.value)}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter') {
                        handleSaveApiKey();
                      }
                    }}
                    className="font-mono"
                  />
                  <Button onClick={handleSaveApiKey} disabled={!inputKey.trim()}>
                    <Check className="h-4 w-4" />
                    저장
                  </Button>
                </div>
              </div>
            )}

            {/* API 키 업데이트 폼 (이미 설정된 경우) */}
            {isAuthenticated && (
              <div className="border-t pt-4">
                <p className="mb-2 text-sm text-muted-foreground">
                  새로운 API 키로 업데이트
                </p>
                <div className="flex gap-2">
                  <Input
                    type="password"
                    placeholder="새 API 키를 입력하세요"
                    value={inputKey}
                    onChange={(e) => setInputKey(e.target.value)}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter') {
                        handleSaveApiKey();
                      }
                    }}
                    className="font-mono"
                  />
                  <Button onClick={handleSaveApiKey} disabled={!inputKey.trim()}>
                    <Check className="h-4 w-4" />
                    업데이트
                  </Button>
                </div>
              </div>
            )}
          </CardContent>
        </Card>

        {/* 연결 상태 카드 */}
        <Card>
          <CardHeader>
            <CardTitle>연결 상태</CardTitle>
            <CardDescription>
              Oracle 데이터베이스 연결 상태를 확인합니다.
            </CardDescription>
          </CardHeader>
          <CardContent>
            {/* 로딩 상태 */}
            {isTestingConnection && (
              <div className="space-y-3">
                <Skeleton className="h-4 w-1/3" />
                <Skeleton className="h-4 w-1/2" />
                <Skeleton className="h-4 w-2/5" />
              </div>
            )}

            {/* 에러 상태 */}
            {connectionError && !isTestingConnection && (
              <Alert variant="destructive">
                <X className="h-4 w-4" />
                <AlertTitle>연결 오류</AlertTitle>
                <AlertDescription>{connectionError}</AlertDescription>
              </Alert>
            )}

            {/* 연결 상태 표시 */}
            {connectionStatus && !isTestingConnection && (
              <div className="space-y-3">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium text-muted-foreground">
                    상태:
                  </span>
                  {connectionStatus.connected ? (
                    <Badge className="gap-1 bg-green-500 hover:bg-green-500/90">
                      <CheckCircle2 className="h-3 w-3" />
                      연결됨
                    </Badge>
                  ) : (
                    <Badge variant="destructive" className="gap-1">
                      <X className="h-3 w-3" />
                      연결 안됨
                    </Badge>
                  )}
                </div>

                {connectionStatus.version && (
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-medium text-muted-foreground">
                      버전:
                    </span>
                    <span className="text-sm">{connectionStatus.version}</span>
                  </div>
                )}

                {connectionStatus.host && (
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-medium text-muted-foreground">
                      호스트:
                    </span>
                    <code className="rounded bg-muted px-1.5 py-0.5 text-sm font-mono">
                      {connectionStatus.host}
                    </code>
                  </div>
                )}

                {connectionStatus.database && (
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-medium text-muted-foreground">
                      데이터베이스:
                    </span>
                    <code className="rounded bg-muted px-1.5 py-0.5 text-sm font-mono">
                      {connectionStatus.database}
                    </code>
                  </div>
                )}

                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium text-muted-foreground">
                    마지막 확인:
                  </span>
                  <span className="text-sm">
                    {new Date(connectionStatus.lastCheck).toLocaleString('ko-KR')}
                  </span>
                </div>

                {connectionStatus.error && (
                  <Alert variant="destructive" className="mt-2">
                    <AlertCircle className="h-4 w-4" />
                    <AlertTitle>오류</AlertTitle>
                    <AlertDescription>{connectionStatus.error}</AlertDescription>
                  </Alert>
                )}
              </div>
            )}

            {/* 테스트 안 한 상태 */}
            {!connectionStatus && !connectionError && !isTestingConnection && (
              <p className="text-sm text-muted-foreground">
                연결 테스트 버튼을 클릭하여 Oracle 데이터베이스 연결 상태를 확인하세요.
              </p>
            )}
          </CardContent>
        </Card>
      </div>
    </MainLayout>
  );
}
