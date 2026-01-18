# Oracle ETL Pipeline

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Oracle DB에서 Google Cloud Storage로 데이터를 추출하는 고성능 ETL 파이프라인 백엔드 API

## 주요 기능

- **Oracle DB 연결 및 데이터 추출**: godror 드라이버를 사용한 고성능 데이터 추출
- **GCS 스트리밍 업로드**: JSONL + Gzip 형식으로 실시간 스트리밍 업로드
- **실시간 진행률 모니터링**: Server-Sent Events (SSE)를 통한 실시간 상태 추적
- **Transport/Job 관리 시스템**: ETL 작업 구성 및 실행 이력 관리
- **병렬 테이블 처리**: 최대 134K rows/sec 처리 성능
- **Graceful Shutdown**: 진행 중인 작업 안전 종료 지원

## 기술 스택

| 분류 | 기술 |
|------|------|
| 언어 | Go 1.23+ |
| 웹 프레임워크 | Fiber v2 |
| Oracle 드라이버 | godror v0.50 |
| GCS SDK | cloud.google.com/go/storage v1.40 |
| 설정 관리 | Viper |
| 로깅 | zerolog (구조화된 JSON 로깅) |
| 인증 | JWT (golang-jwt/jwt/v5) |
| 테스트 | testify |

## 빠른 시작

### 사전 요구사항

- Go 1.23 이상
- Oracle Instant Client (godror 드라이버 요구사항)
- GCP 서비스 계정 (GCS 접근용)

### 설치

```bash
# 저장소 클론
git clone https://github.com/your-org/oracle-etl.git
cd oracle-etl

# 의존성 설치
go mod download

# 빌드
go build -o bin/oracle-etl ./cmd/server
```

### 설정

`config.yaml` 파일 또는 환경변수로 설정할 수 있습니다.

```yaml
# config.yaml
server:
  port: 8080
  read_timeout: 30s
  write_timeout: 300s  # ETL 작업을 위한 긴 타임아웃

app:
  name: oracle-etl
  version: 1.0.0
  environment: development

# Oracle 연결 설정
oracle:
  wallet_path: /opt/wallet
  tns_name: oracle_atp_high
  username: ${ORACLE_USERNAME}
  password: ${ORACLE_PASSWORD}
  pool_min: 5
  pool_max: 20
  fetch_array_size: 1000

# GCS 설정
gcs:
  project_id: ${GCS_PROJECT_ID}
  bucket_name: oracle-etl-data
  credentials_file: /opt/gcp/service-account.json

# ETL 설정
etl:
  chunk_size: 10000
  parallel_tables: 4
  retry_attempts: 3
  retry_backoff: 1s

# 인증 설정
auth:
  enabled: true
  api_keys:
    - ${API_KEY}
  bearer_secret: ${JWT_SECRET}
```

#### 환경변수

| 변수명 | 설명 | 기본값 |
|--------|------|--------|
| `SERVER_PORT` | 서버 포트 | 8080 |
| `ORACLE_WALLET_PATH` | Oracle Wallet 경로 | - |
| `ORACLE_TNS_NAME` | TNS 연결 이름 | - |
| `ORACLE_USERNAME` | Oracle 사용자명 | - |
| `ORACLE_PASSWORD` | Oracle 비밀번호 | - |
| `GCS_PROJECT_ID` | GCP 프로젝트 ID | - |
| `GCS_BUCKET_NAME` | GCS 버킷 이름 | - |
| `GCS_CREDENTIALS_FILE` | 서비스 계정 JSON 경로 | - |
| `AUTH_ENABLED` | 인증 활성화 | false |
| `AUTH_API_KEYS` | API Key 목록 (쉼표 구분) | - |
| `AUTH_BEARER_SECRET` | JWT 서명 비밀키 | - |

### 실행

```bash
# 직접 실행
go run ./cmd/server

# 빌드 후 실행
./bin/oracle-etl

# 설정 파일 지정
CONFIG_PATH=/path/to/config.yaml ./bin/oracle-etl
```

## API 엔드포인트

| 메서드 | 엔드포인트 | 설명 |
|--------|----------|------|
| `GET` | `/api/health` | 서버 상태 확인 |
| `POST` | `/api/transports` | Transport 생성 |
| `GET` | `/api/transports` | Transport 목록 조회 |
| `GET` | `/api/transports/:id` | Transport 상세 조회 |
| `DELETE` | `/api/transports/:id` | Transport 삭제 |
| `POST` | `/api/transports/:id/execute` | Transport 실행 (Job 생성) |
| `GET` | `/api/transports/:id/status` | 실시간 상태 (SSE) |
| `GET` | `/api/jobs` | Job 목록 조회 |
| `GET` | `/api/jobs/:id` | Job 상세 조회 |

자세한 API 문서는 [docs/API.md](docs/API.md)를 참조하세요.

## 사용 예시

### Transport 생성

```bash
curl -X POST http://localhost:8080/api/transports \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{
    "name": "Daily Sales Export",
    "description": "일일 매출 데이터 추출",
    "tables": ["SALES_ORDER", "SALES_LINE_ITEM", "CUSTOMER"]
  }'
```

### Transport 실행

```bash
curl -X POST http://localhost:8080/api/transports/TRPID-abc12345/execute \
  -H "X-API-Key: your-api-key"
```

### 실시간 상태 모니터링 (SSE)

```bash
curl -N http://localhost:8080/api/transports/TRPID-abc12345/status \
  -H "X-API-Key: your-api-key"
```

## 프로젝트 구조

```
oracle-etl/
├── cmd/
│   └── server/           # 서버 진입점
│       └── main.go
├── internal/
│   ├── adapter/          # 외부 시스템 어댑터
│   │   ├── gcs/          # GCS 클라이언트
│   │   ├── handler/      # HTTP 핸들러
│   │   ├── oracle/       # Oracle DB 클라이언트
│   │   └── sse/          # SSE 브로드캐스터
│   ├── config/           # 설정 관리
│   ├── domain/           # 도메인 모델
│   ├── errors/           # 에러 처리
│   ├── middleware/       # HTTP 미들웨어
│   ├── repository/       # 데이터 저장소
│   │   └── memory/       # In-Memory 구현
│   ├── resilience/       # 회복성 패턴
│   └── usecase/          # 비즈니스 로직
├── pkg/
│   ├── buffer/           # 버퍼 관리
│   ├── compress/         # 압축 유틸리티
│   ├── jsonl/            # JSONL 인코딩
│   └── pool/             # 워커 풀
├── benchmarks/           # 벤치마크 테스트
├── tests/
│   ├── e2e/              # E2E 테스트
│   └── integration/      # 통합 테스트
├── config.yaml           # 설정 파일
├── go.mod
└── go.sum
```

자세한 아키텍처 설명은 [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)를 참조하세요.

## 테스트

```bash
# 단위 테스트 실행
go test ./...

# 커버리지 포함 테스트
go test -cover ./...

# 커버리지 리포트 생성
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# 벤치마크 테스트
go test -bench=. ./benchmarks/...
```

## 성능

| 지표 | 수치 |
|------|------|
| 데이터 추출 속도 | 최대 134,000 rows/sec |
| 병렬 테이블 처리 | 최대 4개 동시 처리 |
| 청크 크기 | 10,000 rows (기본값) |
| GCS 업로드 청크 | 16MB (resumable upload) |

## 라이선스

MIT License - 자세한 내용은 [LICENSE](LICENSE) 파일을 참조하세요.

## 기여

프로젝트에 기여하시려면 Pull Request를 제출해주세요.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request
