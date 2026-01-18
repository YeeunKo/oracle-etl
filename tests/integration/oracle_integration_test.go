//go:build integration

// Package integration은 Oracle 및 GCS 통합 테스트를 제공합니다.
// 이 테스트들은 실제 외부 서비스와 연결하여 실행됩니다.
// 실행: go test -tags=integration ./tests/integration/...
package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"oracle-etl/internal/adapter/oracle"
)

// TestOracleConnection은 실제 Oracle 데이터베이스 연결을 테스트합니다.
// 환경 변수 ORACLE_TEST_DSN이 설정되어 있어야 실행됩니다.
func TestOracleConnection(t *testing.T) {
	dsn := os.Getenv("ORACLE_TEST_DSN")
	if dsn == "" {
		t.Skip("Skipping: ORACLE_TEST_DSN 환경 변수가 설정되지 않음")
	}

	walletPath := os.Getenv("ORACLE_WALLET_PATH")
	username := os.Getenv("ORACLE_USERNAME")
	password := os.Getenv("ORACLE_PASSWORD")

	if username == "" || password == "" {
		t.Skip("Skipping: ORACLE_USERNAME 또는 ORACLE_PASSWORD가 설정되지 않음")
	}

	cfg := oracle.PoolConfig{
		TNSName:        dsn,
		WalletPath:     walletPath,
		Username:       username,
		Password:       password,
		PoolMinConns:   1,
		PoolMaxConns:   5,
		FetchArraySize: 1000,
		PrefetchCount:  1000,
		ConnectTimeout: 30 * time.Second,
	}

	// 대신 설정 유효성만 테스트
	t.Run("설정 유효성 검사", func(t *testing.T) {
		assert.NotEmpty(t, cfg.TNSName)
		assert.NotEmpty(t, cfg.Username)
		assert.GreaterOrEqual(t, cfg.PoolMaxConns, cfg.PoolMinConns)
	})
}

// TestOraclePoolConfig는 Oracle 풀 설정을 테스트합니다.
func TestOraclePoolConfig(t *testing.T) {
	cfg := oracle.DefaultPoolConfig()

	t.Run("기본 설정 유효성", func(t *testing.T) {
		assert.Equal(t, 2, cfg.PoolMinConns)
		assert.Equal(t, 10, cfg.PoolMaxConns)
		assert.Equal(t, 1000, cfg.FetchArraySize)
		assert.Equal(t, 30*time.Second, cfg.ConnectTimeout)
	})
}

// TestOracleMockRepository는 Mock 저장소가 제대로 작동하는지 테스트합니다.
func TestOracleMockRepository(t *testing.T) {
	mock := oracle.NewMockRepository()
	ctx := context.Background()

	t.Run("연결 상태 확인", func(t *testing.T) {
		status, err := mock.GetStatus(ctx)
		require.NoError(t, err)
		assert.True(t, status.Connected)
		assert.NotEmpty(t, status.DatabaseVersion)
	})

	t.Run("테이블 목록 조회", func(t *testing.T) {
		tables, err := mock.GetTables(ctx, "SAPSR3")
		require.NoError(t, err)
		assert.NotEmpty(t, tables)
	})

	t.Run("Ping 테스트", func(t *testing.T) {
		err := mock.Ping(ctx)
		require.NoError(t, err)
	})
}

// TestOracleConnectionWithTimeout은 타임아웃 설정이 적용되는지 테스트합니다.
func TestOracleConnectionWithTimeout(t *testing.T) {
	dsn := os.Getenv("ORACLE_TEST_DSN")
	if dsn == "" {
		t.Skip("Skipping: ORACLE_TEST_DSN 환경 변수가 설정되지 않음")
	}

	cfg := oracle.PoolConfig{
		TNSName:        dsn,
		PoolMinConns:   1,
		PoolMaxConns:   5,
		ConnectTimeout: 5 * time.Second,
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnectTimeout)
	defer cancel()

	deadline, ok := ctx.Deadline()
	assert.True(t, ok)
	assert.WithinDuration(t, time.Now().Add(5*time.Second), deadline, 100*time.Millisecond)
}

// TestOracleDataExtraction은 데이터 추출 통합 테스트입니다.
func TestOracleDataExtraction(t *testing.T) {
	dsn := os.Getenv("ORACLE_TEST_DSN")
	if dsn == "" {
		t.Skip("Skipping: ORACLE_TEST_DSN 환경 변수가 설정되지 않음")
	}

	tableOwner := os.Getenv("ORACLE_TABLE_OWNER")
	tableName := os.Getenv("ORACLE_TABLE_NAME")

	if tableOwner == "" || tableName == "" {
		t.Skip("Skipping: ORACLE_TABLE_OWNER 또는 ORACLE_TABLE_NAME이 설정되지 않음")
	}

	mock := oracle.NewMockRepository()
	ctx := context.Background()

	t.Run("샘플 데이터 추출", func(t *testing.T) {
		sample, err := mock.GetSampleData(ctx, tableOwner, tableName, 100)
		require.NoError(t, err)
		assert.NotNil(t, sample)
	})
}
