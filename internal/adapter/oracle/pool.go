// Package oracle은 Oracle 데이터베이스 연결 및 데이터 추출 기능을 제공합니다.
package oracle

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	_ "github.com/godror/godror"
	"oracle-etl/internal/domain"
)

// PoolConfig는 Oracle 커넥션 풀 설정입니다
type PoolConfig struct {
	WalletPath     string        // mTLS wallet 디렉토리 경로
	TNSName        string        // TNS 이름 (예: "oracledb_high")
	Username       string        // Oracle 사용자 이름
	Password       string        // Oracle 비밀번호
	PoolMinConns   int           // 최소 커넥션 수 (기본값: 2)
	PoolMaxConns   int           // 최대 커넥션 수 (기본값: 10)
	FetchArraySize int           // 배치 페치 크기 (기본값: 1000)
	PrefetchCount  int           // 프리페치 카운트 (기본값: 1000)
	ConnectTimeout time.Duration // 연결 타임아웃 (기본값: 30초)
	DefaultOwner   string        // 기본 스키마 소유자
}

// DefaultPoolConfig는 기본 풀 설정을 반환합니다
func DefaultPoolConfig() PoolConfig {
	return PoolConfig{
		PoolMinConns:   2,
		PoolMaxConns:   10,
		FetchArraySize: 1000,
		PrefetchCount:  1000,
		ConnectTimeout: 30 * time.Second,
	}
}

// Pool은 Oracle 커넥션 풀을 관리합니다
type Pool struct {
	db              *sql.DB
	config          PoolConfig
	mu              sync.RWMutex //nolint:unused // 향후 스레드 안전 작업에 사용 예정
	activeConns     int //nolint:unused // 커넥션 모니터링용
	lastCheckedAt   time.Time //nolint:unused // 마지막 체크 시간 기록용
	databaseVersion string //nolint:unused // 캐싱된 버전 정보
	instanceName    string //nolint:unused // 캐싱된 인스턴스 이름
}

// NewPool은 새로운 Oracle 커넥션 풀을 생성합니다
func NewPool(cfg PoolConfig) (*Pool, error) {
	// 설정 기본값 적용
	if cfg.PoolMinConns <= 0 {
		cfg.PoolMinConns = 2
	}
	if cfg.PoolMaxConns <= 0 {
		cfg.PoolMaxConns = 10
	}
	if cfg.FetchArraySize <= 0 {
		cfg.FetchArraySize = 1000
	}
	if cfg.ConnectTimeout <= 0 {
		cfg.ConnectTimeout = 30 * time.Second
	}

	// 연결 문자열 구성
	var connStr string
	if cfg.WalletPath != "" {
		// mTLS wallet 사용 시
		connStr = fmt.Sprintf(`user="%s" password="%s" connectString="%s" 
			poolMinSessions=%d poolMaxSessions=%d 
			configDir="%s"`,
			cfg.Username, cfg.Password, cfg.TNSName,
			cfg.PoolMinConns, cfg.PoolMaxConns,
			cfg.WalletPath)
	} else {
		// 일반 연결
		connStr = fmt.Sprintf(`user="%s" password="%s" connectString="%s" 
			poolMinSessions=%d poolMaxSessions=%d`,
			cfg.Username, cfg.Password, cfg.TNSName,
			cfg.PoolMinConns, cfg.PoolMaxConns)
	}

	db, err := sql.Open("godror", connStr)
	if err != nil {
		return nil, fmt.Errorf("Oracle 연결 실패: %w", err)
	}

	// 연결 풀 설정
	db.SetMaxOpenConns(cfg.PoolMaxConns)
	db.SetMaxIdleConns(cfg.PoolMinConns)
	db.SetConnMaxLifetime(30 * time.Minute)

	pool := &Pool{
		db:     db,
		config: cfg,
	}

	return pool, nil
}

// DB는 내부 sql.DB 인스턴스를 반환합니다 (테스트용)
func (p *Pool) DB() *sql.DB {
	return p.db
}

// Ping은 Oracle 연결을 테스트합니다
func (p *Pool) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

// Close는 커넥션 풀을 종료합니다
func (p *Pool) Close() error {
	return p.db.Close()
}

// GetStatus는 Oracle 연결 상태 및 풀 통계를 반환합니다
func (p *Pool) GetStatus(ctx context.Context) (*domain.OracleStatus, error) {
	status := &domain.OracleStatus{
		CheckedAt: time.Now().UTC(),
	}

	// 연결 테스트
	if err := p.Ping(ctx); err != nil {
		status.Connected = false
		status.Error = err.Error()
		return status, nil
	}

	status.Connected = true

	// 데이터베이스 버전 조회
	var version string
	err := p.db.QueryRowContext(ctx, "SELECT banner FROM v$version WHERE ROWNUM = 1").Scan(&version)
	if err != nil {
		// 버전 조회 실패해도 연결은 성공한 것으로 처리
		version = "Unknown"
	}
	status.DatabaseVersion = extractVersion(version)

	// 인스턴스 이름 조회
	var instanceName string
	err = p.db.QueryRowContext(ctx, "SELECT instance_name FROM v$instance").Scan(&instanceName)
	if err != nil {
		instanceName = "Unknown"
	}
	status.InstanceName = instanceName

	// 풀 통계 조회
	stats := p.db.Stats()
	status.PoolStats = domain.PoolStats{
		ActiveConnections: stats.InUse,
		IdleConnections:   stats.Idle,
		MaxConnections:    p.config.PoolMaxConns,
	}

	return status, nil
}

// GetTables는 접근 가능한 테이블 목록과 row count를 반환합니다
func (p *Pool) GetTables(ctx context.Context, owner string) ([]domain.TableInfo, error) {
	query := `
		SELECT 
			t.table_name,
			t.owner,
			NVL(t.num_rows, 0) as row_count,
			(SELECT COUNT(*) FROM all_tab_columns c WHERE c.owner = t.owner AND c.table_name = t.table_name) as column_count
		FROM all_tables t
		WHERE t.owner = :1
		ORDER BY t.table_name
	`

	rows, err := p.db.QueryContext(ctx, query, owner)
	if err != nil {
		return nil, fmt.Errorf("테이블 목록 조회 실패: %w", err)
	}
	defer rows.Close()

	var tables []domain.TableInfo
	for rows.Next() {
		var t domain.TableInfo
		if err := rows.Scan(&t.Name, &t.Owner, &t.RowCount, &t.ColumnCount); err != nil {
			return nil, fmt.Errorf("테이블 정보 스캔 실패: %w", err)
		}
		tables = append(tables, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("테이블 목록 순회 실패: %w", err)
	}

	return tables, nil
}

// GetTableColumns는 테이블의 컬럼 정보를 반환합니다
func (p *Pool) GetTableColumns(ctx context.Context, owner, tableName string) ([]domain.ColumnInfo, error) {
	query := `
		SELECT 
			column_name,
			data_type,
			CASE WHEN nullable = 'Y' THEN 1 ELSE 0 END as nullable,
			column_id
		FROM all_tab_columns
		WHERE owner = :1 AND table_name = :2
		ORDER BY column_id
	`

	rows, err := p.db.QueryContext(ctx, query, owner, tableName)
	if err != nil {
		return nil, fmt.Errorf("컬럼 목록 조회 실패: %w", err)
	}
	defer rows.Close()

	var columns []domain.ColumnInfo
	for rows.Next() {
		var c domain.ColumnInfo
		var nullable int
		if err := rows.Scan(&c.Name, &c.DataType, &nullable, &c.Position); err != nil {
			return nil, fmt.Errorf("컬럼 정보 스캔 실패: %w", err)
		}
		c.Nullable = nullable == 1
		columns = append(columns, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("컬럼 목록 순회 실패: %w", err)
	}

	return columns, nil
}

// GetSampleData는 테이블의 샘플 데이터를 반환합니다
func (p *Pool) GetSampleData(ctx context.Context, owner, tableName string, limit int) (*domain.SampleData, error) {
	if limit <= 0 {
		limit = 100
	}

	// 컬럼 정보 조회
	columns, err := p.GetTableColumns(ctx, owner, tableName)
	if err != nil {
		return nil, err
	}

	columnNames := make([]string, len(columns))
	for i, c := range columns {
		columnNames[i] = c.Name
	}

	// 샘플 데이터 조회
	// #nosec G201 -- owner와 tableName은 API 레벨에서 검증된 입력값입니다
	query := fmt.Sprintf("SELECT * FROM %s.%s WHERE ROWNUM <= :1", owner, tableName)
	rows, err := p.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("샘플 데이터 조회 실패: %w", err)
	}
	defer rows.Close()

	// 결과 스캔
	data := &domain.SampleData{
		TableName: tableName,
		Columns:   columnNames,
		Rows:      []map[string]interface{}{},
	}

	colTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, fmt.Errorf("컬럼 타입 조회 실패: %w", err)
	}

	for rows.Next() {
		// 동적 스캔을 위한 슬라이스 생성
		values := make([]interface{}, len(colTypes))
		valuePtrs := make([]interface{}, len(colTypes))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("데이터 스캔 실패: %w", err)
		}

		row := make(map[string]interface{})
		for i, col := range columnNames {
			row[col] = convertValue(values[i])
		}
		data.Rows = append(data.Rows, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("데이터 순회 실패: %w", err)
	}

	data.Count = len(data.Rows)
	return data, nil
}

// StreamTableData는 테이블 데이터를 청크 단위로 스트리밍합니다
func (p *Pool) StreamTableData(ctx context.Context, owner, tableName string, opts domain.ExtractionOptions, chunkHandler func(chunk *domain.ChunkResult) error) error {
	if opts.ChunkSize <= 0 {
		opts.ChunkSize = 10000
	}
	if opts.FetchArraySize <= 0 {
		opts.FetchArraySize = p.config.FetchArraySize
	}

	// 전체 데이터 조회 쿼리
	// #nosec G201 -- owner와 tableName은 API 레벨에서 검증된 입력값입니다
	query := fmt.Sprintf("SELECT * FROM %s.%s", owner, tableName)
	rows, err := p.db.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("데이터 스트리밍 시작 실패: %w", err)
	}
	defer rows.Close()

	// 컬럼 정보 조회
	colTypes, err := rows.ColumnTypes()
	if err != nil {
		return fmt.Errorf("컬럼 타입 조회 실패: %w", err)
	}

	columnNames := make([]string, len(colTypes))
	for i, ct := range colTypes {
		columnNames[i] = ct.Name()
	}

	chunkNumber := 0
	var totalRowsSent int64
	chunkRows := make([]map[string]interface{}, 0, opts.ChunkSize)

	for rows.Next() {
		// 동적 스캔
		values := make([]interface{}, len(colTypes))
		valuePtrs := make([]interface{}, len(colTypes))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return fmt.Errorf("데이터 스캔 실패: %w", err)
		}

		row := make(map[string]interface{})
		for i, col := range columnNames {
			row[col] = convertValue(values[i])
		}
		chunkRows = append(chunkRows, row)

		// 청크가 가득 찼으면 핸들러 호출
		if len(chunkRows) >= opts.ChunkSize {
			chunkNumber++
			totalRowsSent += int64(len(chunkRows))
			chunk := &domain.ChunkResult{
				TableName:     tableName,
				ChunkNumber:   chunkNumber,
				Rows:          chunkRows,
				RowCount:      len(chunkRows),
				IsLastChunk:   false,
				TotalRowsSent: totalRowsSent,
			}
			if err := chunkHandler(chunk); err != nil {
				return fmt.Errorf("청크 핸들러 오류: %w", err)
			}
			chunkRows = make([]map[string]interface{}, 0, opts.ChunkSize)
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("데이터 순회 실패: %w", err)
	}

	// 마지막 청크 처리
	if len(chunkRows) > 0 {
		chunkNumber++
		totalRowsSent += int64(len(chunkRows))
		chunk := &domain.ChunkResult{
			TableName:     tableName,
			ChunkNumber:   chunkNumber,
			Rows:          chunkRows,
			RowCount:      len(chunkRows),
			IsLastChunk:   true,
			TotalRowsSent: totalRowsSent,
		}
		if err := chunkHandler(chunk); err != nil {
			return fmt.Errorf("마지막 청크 핸들러 오류: %w", err)
		}
	}

	return nil
}

// extractVersion은 Oracle 버전 문자열에서 버전 번호를 추출합니다
func extractVersion(banner string) string {
	// 예: "Oracle Database 19c Enterprise Edition Release 19.0.0.0.0"
	parts := strings.Fields(banner)
	for _, part := range parts {
		if strings.Contains(part, ".") && strings.Count(part, ".") >= 2 {
			return part
		}
	}
	return banner
}

// convertValue는 Oracle 값을 JSON 직렬화 가능한 형태로 변환합니다
func convertValue(v interface{}) interface{} {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case []byte:
		return string(val)
	case time.Time:
		return val.Format(time.RFC3339)
	default:
		return val
	}
}

// 인터페이스 구현 확인
var _ Repository = (*Pool)(nil)
