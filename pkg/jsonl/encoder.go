// Package jsonl은 JSON Lines (NDJSON) 인코딩 기능을 제공합니다.
// JSON Lines 형식은 각 줄에 하나의 JSON 객체를 포함하며,
// 스트리밍 처리와 대용량 데이터 처리에 적합합니다.
package jsonl

import (
	"bufio"
	"encoding/json"
	"io"
	"sync/atomic"
)

// Encoder는 JSON Lines 형식으로 데이터를 인코딩하는 인터페이스입니다
type Encoder interface {
	// Encode는 단일 객체를 JSON Lines 형식으로 인코딩합니다
	Encode(v interface{}) error

	// Flush는 버퍼에 남은 데이터를 출력합니다
	Flush() error

	// BytesWritten은 지금까지 기록된 총 바이트 수를 반환합니다
	BytesWritten() int64

	// RowsEncoded는 지금까지 인코딩된 row 수를 반환합니다
	RowsEncoded() int64
}

// encoder는 Encoder 인터페이스의 기본 구현체입니다
type encoder struct {
	writer       *bufio.Writer
	jsonEncoder  *json.Encoder
	bytesWritten int64
	rowsEncoded  int64
	countWriter  *countingWriter
}

// countingWriter는 기록된 바이트 수를 추적하는 io.Writer 래퍼입니다
type countingWriter struct {
	writer       io.Writer
	bytesWritten *int64
}

// Write는 io.Writer 인터페이스를 구현합니다
func (cw *countingWriter) Write(p []byte) (n int, err error) {
	n, err = cw.writer.Write(p)
	atomic.AddInt64(cw.bytesWritten, int64(n))
	return
}

// NewEncoder는 새로운 JSON Lines 인코더를 생성합니다
// writer: 인코딩된 데이터를 기록할 io.Writer
func NewEncoder(w io.Writer) Encoder {
	var bytesWritten int64
	countWriter := &countingWriter{
		writer:       w,
		bytesWritten: &bytesWritten,
	}

	bufWriter := bufio.NewWriterSize(countWriter, 64*1024) // 64KB 버퍼
	jsonEnc := json.NewEncoder(bufWriter)
	jsonEnc.SetEscapeHTML(false) // HTML 이스케이프 비활성화 (성능 향상)

	return &encoder{
		writer:       bufWriter,
		jsonEncoder:  jsonEnc,
		bytesWritten: 0,
		rowsEncoded:  0,
		countWriter:  countWriter,
	}
}

// Encode는 단일 객체를 JSON Lines 형식으로 인코딩합니다
// JSON Lines 형식: 각 객체는 단일 줄에 JSON으로 직렬화되고 개행문자로 구분됩니다
func (e *encoder) Encode(v interface{}) error {
	if err := e.jsonEncoder.Encode(v); err != nil {
		return err
	}
	atomic.AddInt64(&e.rowsEncoded, 1)
	return nil
}

// Flush는 버퍼에 남은 데이터를 출력합니다
func (e *encoder) Flush() error {
	return e.writer.Flush()
}

// BytesWritten은 지금까지 기록된 총 바이트 수를 반환합니다
func (e *encoder) BytesWritten() int64 {
	// 버퍼에 아직 있는 데이터도 포함하여 계산
	_ = e.writer.Flush() // 바이트 계산용으로 에러 무시
	return atomic.LoadInt64(e.countWriter.bytesWritten)
}

// RowsEncoded는 지금까지 인코딩된 row 수를 반환합니다
func (e *encoder) RowsEncoded() int64 {
	return atomic.LoadInt64(&e.rowsEncoded)
}

// StreamEncoder는 스트리밍 컨텍스트에서 사용하기 위한 확장 인터페이스입니다
type StreamEncoder interface {
	Encoder

	// EncodeRows는 여러 row를 한 번에 인코딩합니다
	EncodeRows(rows []map[string]interface{}) error

	// Reset은 내부 상태를 초기화합니다 (동일한 writer에서 재사용)
	Reset()
}

// streamEncoder는 StreamEncoder 인터페이스의 구현체입니다
type streamEncoder struct {
	*encoder
}

// NewStreamEncoder는 스트리밍용 JSON Lines 인코더를 생성합니다
func NewStreamEncoder(w io.Writer) StreamEncoder {
	enc := NewEncoder(w).(*encoder)
	return &streamEncoder{encoder: enc}
}

// EncodeRows는 여러 row를 한 번에 인코딩합니다
func (e *streamEncoder) EncodeRows(rows []map[string]interface{}) error {
	for _, row := range rows {
		if err := e.Encode(row); err != nil {
			return err
		}
	}
	return nil
}

// Reset은 카운터를 초기화합니다
func (e *streamEncoder) Reset() {
	atomic.StoreInt64(&e.rowsEncoded, 0)
	atomic.StoreInt64(e.countWriter.bytesWritten, 0)
}
