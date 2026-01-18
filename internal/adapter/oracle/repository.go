// Package oracle은 Oracle 데이터베이스 연결 및 데이터 추출 기능을 제공합니다.
package oracle

import (
	"context"

	"oracle-etl/internal/domain"
)

// Repository는 Oracle 데이터베이스 작업을 위한 인터페이스입니다.
// 테스트 용이성을 위해 인터페이스로 정의합니다.
type Repository interface {
	// GetStatus는 Oracle 연결 상태 및 풀 통계를 반환합니다
	GetStatus(ctx context.Context) (*domain.OracleStatus, error)

	// GetTables는 접근 가능한 테이블 목록과 row count를 반환합니다
	GetTables(ctx context.Context, owner string) ([]domain.TableInfo, error)

	// GetTableColumns는 테이블의 컬럼 정보를 반환합니다
	GetTableColumns(ctx context.Context, owner, tableName string) ([]domain.ColumnInfo, error)

	// GetSampleData는 테이블의 샘플 데이터를 반환합니다
	GetSampleData(ctx context.Context, owner, tableName string, limit int) (*domain.SampleData, error)

	// StreamTableData는 테이블 데이터를 청크 단위로 스트리밍합니다
	StreamTableData(ctx context.Context, owner, tableName string, opts domain.ExtractionOptions, chunkHandler func(chunk *domain.ChunkResult) error) error

	// Ping은 Oracle 연결을 테스트합니다
	Ping(ctx context.Context) error

	// Close는 커넥션 풀을 종료합니다
	Close() error
}
