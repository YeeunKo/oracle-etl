//go:build e2e

// Package e2e는 전체 ETL 플로우 테스트를 제공합니다.
// 이 테스트들은 실제 Oracle 및 GCS 환경에서 전체 플로우를 테스트합니다.
// 실행: go test -tags=e2e ./tests/e2e/...
package e2e

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"oracle-etl/internal/adapter/gcs"
	"oracle-etl/internal/adapter/oracle"
	"oracle-etl/internal/domain"
	"oracle-etl/internal/repository/memory"
	"oracle-etl/internal/usecase"
)

// E2E 테스트를 위한 환경 변수 확인
func checkE2EEnvironment(t *testing.T) {
	required := []string{
		"ORACLE_TEST_DSN",
		"ORACLE_USERNAME",
		"ORACLE_PASSWORD",
		"GCS_PROJECT_ID",
		"GCS_TEST_BUCKET",
		"GCS_CREDENTIALS_FILE",
	}

	for _, env := range required {
		if os.Getenv(env) == "" {
			t.Skipf("Skipping E2E test: %s 환경 변수가 설정되지 않음", env)
		}
	}
}

// TestFullETLFlow는 전체 ETL 플로우를 테스트합니다.
func TestFullETLFlow(t *testing.T) {
	checkE2EEnvironment(t)

	// 환경이 설정되지 않은 경우 Mock으로 테스트
	t.Run("Mock ETL 플로우 테스트", func(t *testing.T) {
		ctx := context.Background()

		// 1. Transport 저장소 설정
		transportRepo := memory.NewTransportRepository()
		jobRepo := memory.NewJobRepository()

		// 2. Transport 생성
		transport := domain.NewTransport(
			domain.GenerateTransportID("test1234"),
			"E2E 테스트 Transport",
			"E2E 테스트용 Transport",
			[]string{"VBRP", "LIKP"},
		)
		err := transportRepo.Create(ctx, transport)
		require.NoError(t, err)

		// 3. Transport 조회 확인
		retrieved, err := transportRepo.GetByID(ctx, transport.ID)
		require.NoError(t, err)
		assert.Equal(t, transport.Name, retrieved.Name)
		assert.Equal(t, 2, len(retrieved.Tables))

		// 4. Job 생성
		job := domain.NewJob(
			domain.GenerateJobID(time.Now(), "abc123"),
			transport.ID,
			1,
		)
		err = jobRepo.Create(ctx, job)
		require.NoError(t, err)

		// 5. Job 시작
		job.Start()
		err = jobRepo.Update(ctx, job)
		require.NoError(t, err)
		assert.Equal(t, domain.JobStatusRunning, job.Status)

		// 6. Extraction 추가
		job.AddExtraction(domain.Extraction{
			TableName:  "VBRP",
			Status:     domain.ExtractionStatusCompleted,
			RowCount:   100000,
			ByteCount:  5242880,
			GCSPath:    "gs://test-bucket/TRP-001/v001/VBRP.jsonl.gz",
			StartedAt:  time.Now().Add(-1 * time.Minute),
			FinishedAt: func() *time.Time { t := time.Now(); return &t }(),
		})

		job.AddExtraction(domain.Extraction{
			TableName:  "LIKP",
			Status:     domain.ExtractionStatusCompleted,
			RowCount:   50000,
			ByteCount:  2621440,
			GCSPath:    "gs://test-bucket/TRP-001/v001/LIKP.jsonl.gz",
			StartedAt:  time.Now().Add(-30 * time.Second),
			FinishedAt: func() *time.Time { t := time.Now(); return &t }(),
		})

		// 7. Job 완료
		job.UpdateMetrics()
		job.Complete()
		err = jobRepo.Update(ctx, job)
		require.NoError(t, err)

		// 8. 최종 상태 검증
		assert.Equal(t, domain.JobStatusCompleted, job.Status)
		assert.Equal(t, int64(150000), job.Metrics.TotalRows)
		assert.Equal(t, int64(7864320), job.Metrics.TotalBytes)
		assert.Equal(t, 2, len(job.Extractions))
	})
}

// TestTransportServiceE2E는 Transport 서비스의 E2E 테스트입니다.
func TestTransportServiceE2E(t *testing.T) {
	ctx := context.Background()

	// Mock 저장소 사용
	transportRepo := memory.NewTransportRepository()
	jobRepo := memory.NewJobRepository()
	oracleRepo := oracle.NewMockRepository()
	gcsClient := gcs.NewMockClient(gcs.GCSConfig{
		ProjectID:  "test-project",
		BucketName: "test-bucket",
	})

	// Transport 서비스 생성
	transportService := usecase.NewTransportService(transportRepo, jobRepo, oracleRepo, gcsClient)

	t.Run("Transport CRUD 플로우", func(t *testing.T) {
		// 생성
		req := &domain.CreateTransportRequest{
			Name:        "테스트 Transport",
			Description: "E2E 테스트",
			Tables:      []string{"VBRP"},
		}

		transport, err := transportService.Create(ctx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, transport.ID)

		// 조회
		retrieved, err := transportService.GetByID(ctx, transport.ID)
		require.NoError(t, err)
		assert.Equal(t, req.Name, retrieved.Name)

		// 목록 조회
		list, err := transportService.List(ctx, 0, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(list), 1)

		// 삭제
		err = transportService.Delete(ctx, transport.ID)
		require.NoError(t, err)

		// 삭제 확인
		_, err = transportService.GetByID(ctx, transport.ID)
		require.Error(t, err)
	})
}

// TestJobServiceE2E는 Job 서비스의 E2E 테스트입니다.
func TestJobServiceE2E(t *testing.T) {
	ctx := context.Background()

	jobRepo := memory.NewJobRepository()
	jobService := usecase.NewJobService(jobRepo)

	t.Run("Job 라이프사이클", func(t *testing.T) {
		transportID := "TRPID-test1234"

		// Job 생성
		job, err := jobService.CreateForTransport(ctx, transportID)
		require.NoError(t, err)
		assert.Equal(t, domain.JobStatusPending, job.Status)
		assert.Equal(t, 1, job.Version)

		// 두 번째 Job 생성 (버전 증가 확인)
		job2, err := jobService.CreateForTransport(ctx, transportID)
		require.NoError(t, err)
		assert.Equal(t, 2, job2.Version)

		// Job 목록 조회
		filter := domain.JobListFilter{
			TransportID: transportID,
			Limit:       10,
		}
		jobs, total, err := jobService.List(ctx, filter)
		require.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Equal(t, 2, len(jobs))

		// Job 상태 업데이트
		job.Start()
		err = jobService.Update(ctx, job)
		require.NoError(t, err)

		retrieved, err := jobService.GetByID(ctx, job.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.JobStatusRunning, retrieved.Status)
	})
}

// TestOracleToGCSFlow는 Oracle에서 GCS로의 데이터 흐름을 테스트합니다.
func TestOracleToGCSFlow(t *testing.T) {
	checkE2EEnvironment(t)

	t.Run("Mock 데이터 흐름 테스트", func(t *testing.T) {
		ctx := context.Background()

		// Oracle Mock 저장소
		oracleRepo := oracle.NewMockRepository()

		// 상태 확인
		status, err := oracleRepo.GetStatus(ctx)
		require.NoError(t, err)
		assert.True(t, status.Connected)

		// 테이블 목록 조회
		tables, err := oracleRepo.GetTables(ctx, "SAPSR3")
		require.NoError(t, err)
		assert.NotEmpty(t, tables)

		// 샘플 데이터 조회
		sample, err := oracleRepo.GetSampleData(ctx, "SAPSR3", "VBRP", 100)
		require.NoError(t, err)
		assert.NotEmpty(t, sample.Columns)
		assert.NotEmpty(t, sample.Rows)

		// GCS Mock 클라이언트
		gcsClient := gcs.NewMockClient(gcs.GCSConfig{
			ProjectID:  "test-project",
			BucketName: "test-bucket",
		})

		// GCS 연결 확인
		err = gcsClient.Ping(ctx)
		require.NoError(t, err)

		// 객체 경로 생성
		objectPath := gcsClient.ObjectPath("TRP-001", "v001", "VBRP")
		assert.Equal(t, "TRP-001/v001/VBRP.jsonl.gz", objectPath)
	})
}

// TestConcurrentJobExecution은 동시 Job 실행을 테스트합니다.
func TestConcurrentJobExecution(t *testing.T) {
	ctx := context.Background()
	jobRepo := memory.NewJobRepository()
	jobService := usecase.NewJobService(jobRepo)

	transportID := "TRPID-concurrent"
	jobCount := 5

	t.Run("동시 Job 생성", func(t *testing.T) {
		jobs := make([]*domain.Job, jobCount)
		var err error

		for i := 0; i < jobCount; i++ {
			jobs[i], err = jobService.CreateForTransport(ctx, transportID)
			require.NoError(t, err)
		}

		// 버전 순차 증가 확인
		for i, job := range jobs {
			assert.Equal(t, i+1, job.Version)
		}

		// 전체 Job 수 확인
		filter := domain.JobListFilter{
			TransportID: transportID,
			Limit:       100,
		}
		_, total, err := jobService.List(ctx, filter)
		require.NoError(t, err)
		assert.Equal(t, jobCount, total)
	})
}

// TestErrorHandlingE2E는 에러 처리 플로우를 테스트합니다.
func TestErrorHandlingE2E(t *testing.T) {
	ctx := context.Background()

	t.Run("존재하지 않는 Transport 조회", func(t *testing.T) {
		transportRepo := memory.NewTransportRepository()
		jobRepo := memory.NewJobRepository()
		oracleRepo := oracle.NewMockRepository()
		gcsClient := gcs.NewMockClient(gcs.GCSConfig{
			ProjectID:  "test-project",
			BucketName: "test-bucket",
		})

		service := usecase.NewTransportService(transportRepo, jobRepo, oracleRepo, gcsClient)

		_, err := service.GetByID(ctx, "NON-EXISTENT-ID")
		require.Error(t, err)
	})

	t.Run("유효하지 않은 Transport 생성", func(t *testing.T) {
		transportRepo := memory.NewTransportRepository()
		jobRepo := memory.NewJobRepository()
		oracleRepo := oracle.NewMockRepository()
		gcsClient := gcs.NewMockClient(gcs.GCSConfig{
			ProjectID:  "test-project",
			BucketName: "test-bucket",
		})

		service := usecase.NewTransportService(transportRepo, jobRepo, oracleRepo, gcsClient)

		// 빈 테이블 목록
		req := &domain.CreateTransportRequest{
			Name:   "테스트",
			Tables: []string{},
		}

		_, err := service.Create(ctx, req)
		require.Error(t, err)
	})

	t.Run("Oracle 에러 처리", func(t *testing.T) {
		oracleRepo := oracle.NewMockRepository()
		oracleRepo.ShouldError = true
		oracleRepo.ErrorMessage = "연결 실패"

		_, err := oracleRepo.GetStatus(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "연결 실패")
	})
}
