/**
 * shadcn/ui 유틸리티 함수
 */

import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

/**
 * 클래스명을 조건부로 병합하는 유틸리티 함수
 * clsx와 tailwind-merge를 조합하여 Tailwind CSS 클래스 충돌을 해결
 *
 * @param inputs - 클래스명 배열 (조건부 클래스 포함 가능)
 * @returns 병합된 클래스명 문자열
 *
 * @example
 * cn("px-4 py-2", isActive && "bg-blue-500", "text-white")
 * // => "px-4 py-2 bg-blue-500 text-white" (isActive가 true일 때)
 */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}
