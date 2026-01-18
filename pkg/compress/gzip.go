// Package compress는 데이터 압축 기능을 제공합니다.
// GCS 업로드를 위한 gzip 스트리밍 압축을 지원합니다.
package compress

import (
	"compress/gzip"
	"io"
	"sync/atomic"
)

// GzipWriter는 gzip 압축을 수행하는 io.WriteCloser 래퍼입니다
type GzipWriter interface {
	io.WriteCloser

	// Flush는 버퍼에 남은 데이터를 출력합니다
	Flush() error

	// BytesWritten은 압축 후 기록된 바이트 수를 반환합니다
	BytesWritten() int64

	// BytesRead는 압축 전 입력된 바이트 수를 반환합니다
	BytesRead() int64
}

// gzipWriter는 GzipWriter 인터페이스의 구현체입니다
type gzipWriter struct {
	writer       *gzip.Writer
	countWriter  *countingWriter
	bytesRead    int64
}

// countingWriter는 기록된 바이트 수를 추적하는 io.Writer 래퍼입니다
type countingWriter struct {
	writer       io.Writer
	bytesWritten int64
}

// Write는 io.Writer 인터페이스를 구현합니다
func (cw *countingWriter) Write(p []byte) (n int, err error) {
	n, err = cw.writer.Write(p)
	atomic.AddInt64(&cw.bytesWritten, int64(n))
	return
}

// NewGzipWriter는 새로운 gzip writer를 생성합니다
// 기본 압축 레벨 (BestSpeed와 BestCompression 사이)을 사용합니다
func NewGzipWriter(w io.Writer) GzipWriter {
	return NewGzipWriterLevel(w, gzip.DefaultCompression)
}

// NewGzipWriterLevel은 지정된 압축 레벨로 gzip writer를 생성합니다
// level: gzip.NoCompression, gzip.BestSpeed, gzip.BestCompression, gzip.DefaultCompression
func NewGzipWriterLevel(w io.Writer, level int) GzipWriter {
	countWriter := &countingWriter{
		writer:       w,
		bytesWritten: 0,
	}

	gzWriter, _ := gzip.NewWriterLevel(countWriter, level)

	return &gzipWriter{
		writer:      gzWriter,
		countWriter: countWriter,
		bytesRead:   0,
	}
}

// Write는 데이터를 압축하여 기록합니다
func (g *gzipWriter) Write(p []byte) (n int, err error) {
	n, err = g.writer.Write(p)
	atomic.AddInt64(&g.bytesRead, int64(n))
	return
}

// Flush는 버퍼에 남은 데이터를 출력합니다
func (g *gzipWriter) Flush() error {
	return g.writer.Flush()
}

// Close는 gzip writer를 닫습니다
// 반드시 호출해야 압축이 완료됩니다
func (g *gzipWriter) Close() error {
	return g.writer.Close()
}

// BytesWritten은 압축 후 기록된 바이트 수를 반환합니다
func (g *gzipWriter) BytesWritten() int64 {
	return atomic.LoadInt64(&g.countWriter.bytesWritten)
}

// BytesRead는 압축 전 입력된 바이트 수를 반환합니다
func (g *gzipWriter) BytesRead() int64 {
	return atomic.LoadInt64(&g.bytesRead)
}

// GzipReader는 gzip 압축 해제를 위한 래퍼입니다 (테스트용)
type GzipReader struct {
	reader *gzip.Reader
}

// NewGzipReader는 새로운 gzip reader를 생성합니다
func NewGzipReader(r io.Reader) (*GzipReader, error) {
	gzReader, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return &GzipReader{reader: gzReader}, nil
}

// Read는 io.Reader 인터페이스를 구현합니다
func (g *GzipReader) Read(p []byte) (n int, err error) {
	return g.reader.Read(p)
}

// Close는 gzip reader를 닫습니다
func (g *GzipReader) Close() error {
	return g.reader.Close()
}
