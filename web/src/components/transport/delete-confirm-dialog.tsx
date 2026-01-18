/**
 * Transport 삭제 확인 다이얼로그
 * FR-02.6: 삭제 전 확인
 * FR-02.7: 실행 중인 Transport 삭제 방지
 */

'use client';

import { AlertTriangle } from 'lucide-react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';

// ============================================================================
// Props 타입
// ============================================================================

interface DeleteConfirmDialogProps {
  /** 다이얼로그 열림 상태 */
  open: boolean;
  /** 열림 상태 변경 핸들러 */
  onOpenChange: (open: boolean) => void;
  /** 삭제할 Transport 이름 */
  transportName: string;
  /** 삭제 확인 핸들러 */
  onConfirm: () => void;
  /** 삭제 진행 중 여부 */
  isDeleting?: boolean;
}

// ============================================================================
// 메인 컴포넌트
// ============================================================================

/**
 * Transport 삭제 확인 다이얼로그
 * 삭제 전 사용자에게 확인을 요청
 *
 * @example
 * <DeleteConfirmDialog
 *   open={showDialog}
 *   onOpenChange={setShowDialog}
 *   transportName="Daily Sales Export"
 *   onConfirm={handleDelete}
 *   isDeleting={isDeleting}
 * />
 */
export function DeleteConfirmDialog({
  open,
  onOpenChange,
  transportName,
  onConfirm,
  isDeleting = false,
}: DeleteConfirmDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-full bg-destructive/10">
              <AlertTriangle className="h-5 w-5 text-destructive" />
            </div>
            <DialogTitle>Transport 삭제</DialogTitle>
          </div>
          <DialogDescription className="pt-2">
            <span className="font-semibold text-foreground">
              &quot;{transportName}&quot;
            </span>
            을(를) 삭제하시겠습니까?
          </DialogDescription>
        </DialogHeader>

        <div className="rounded-md border border-destructive/20 bg-destructive/5 p-3">
          <p className="text-sm text-destructive">
            이 작업은 되돌릴 수 없습니다. Transport 설정과 관련된 실행 기록이 모두 삭제됩니다.
          </p>
        </div>

        <DialogFooter className="gap-2 sm:gap-0">
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={isDeleting}
          >
            취소
          </Button>
          <Button
            variant="destructive"
            onClick={onConfirm}
            disabled={isDeleting}
          >
            {isDeleting ? '삭제 중...' : '삭제'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
