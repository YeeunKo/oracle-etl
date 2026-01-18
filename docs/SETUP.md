# 환경 설정 가이드

Oracle ETL Pipeline의 개발 및 운영 환경 설정을 위한 상세 가이드입니다.

## 목차

- [사전 요구사항](#사전-요구사항)
- [Oracle 환경 설정](#oracle-환경-설정)
- [GCS 환경 설정](#gcs-환경-설정)
- [애플리케이션 설정](#애플리케이션-설정)
- [프론트엔드 설정](#프론트엔드-설정)
- [Docker 환경](#docker-환경)
- [문제 해결](#문제-해결)

---

## 사전 요구사항

### 백엔드

| 요구사항 | 버전 | 설치 확인 |
|----------|------|----------|
| Go | 1.23+ | `go version` |
| Oracle Instant Client | 21.x+ | `ls $ORACLE_HOME` |
| GCP CLI (선택) | latest | `gcloud --version` |

### 프론트엔드

| 요구사항 | 버전 | 설치 확인 |
|----------|------|----------|
| Node.js | 20+ | `node --version` |
| npm | 10+ | `npm --version` |

---

## Oracle 환경 설정

### 1. Oracle Instant Client 설치

#### macOS (Homebrew)

```bash
# Homebrew로 설치
brew tap InstantClientTap/instantclient
brew install instantclient-basic
brew install instantclient-sdk

# 환경변수 설정 (~/.zshrc 또는 ~/.bashrc)
export ORACLE_HOME=$(brew --prefix)/lib
export DYLD_LIBRARY_PATH=$ORACLE_HOME:$DYLD_LIBRARY_PATH
```

#### macOS (수동 설치)

```bash
# Oracle 공식 사이트에서 다운로드
# https://www.oracle.com/database/technologies/instant-client/macos-intel-x86-downloads.html

# 압축 해제 및 이동
mkdir -p /opt/oracle
unzip instantclient-basic-macos.x64-21.x.x.x.x.zip -d /opt/oracle
unzip instantclient-sdk-macos.x64-21.x.x.x.x.zip -d /opt/oracle

# 환경변수 설정
export ORACLE_HOME=/opt/oracle/instantclient_21_x
export DYLD_LIBRARY_PATH=$ORACLE_HOME:$DYLD_LIBRARY_PATH
```

#### Linux (Ubuntu/Debian)

```bash
# Oracle Instant Client RPM 다운로드 후 설치
# https://www.oracle.com/database/technologies/instant-client/linux-x86-64-downloads.html

# Alien으로 RPM을 DEB로 변환
sudo apt-get install alien libaio1
sudo alien -i oracle-instantclient21.x-basic-21.x.x.x.x-1.x86_64.rpm
sudo alien -i oracle-instantclient21.x-devel-21.x.x.x.x-1.x86_64.rpm

# 환경변수 설정
export ORACLE_HOME=/usr/lib/oracle/21/client64
export LD_LIBRARY_PATH=$ORACLE_HOME/lib:$LD_LIBRARY_PATH
```

#### Windows

```powershell
# Oracle 공식 사이트에서 다운로드
# https://www.oracle.com/database/technologies/instant-client/winx64-64-downloads.html

# 압축 해제 후 PATH에 추가
# 시스템 환경변수에서 설정:
# ORACLE_HOME = C:\oracle\instantclient_21_x
# PATH에 %ORACLE_HOME% 추가
```

### 2. Oracle Wallet 설정 (Oracle Autonomous Database용)

Oracle Autonomous Database (ATP/ADW)에 연결하려면 Wallet이 필요합니다.

```bash
# 1. Oracle Cloud Console에서 Wallet 다운로드
#    - Database Details → DB Connection → Download Wallet
#    - 암호 설정 후 Wallet_xxx.zip 다운로드

# 2. Wallet 압축 해제
mkdir -p /opt/oracle/wallet
unzip Wallet_xxx.zip -d /opt/oracle/wallet

# 3. sqlnet.ora 수정 (Wallet 경로 설정)
# /opt/oracle/wallet/sqlnet.ora 파일 편집:
# WALLET_LOCATION = (SOURCE = (METHOD = file) (METHOD_DATA = (DIRECTORY="/opt/oracle/wallet")))
# SSL_SERVER_DN_MATCH=yes

# 4. 환경변수 설정
export TNS_ADMIN=/opt/oracle/wallet
```

### 3. Oracle 연결 테스트

```bash
# Go 테스트 코드 실행
cd oracle-etl
go test ./tests/integration/oracle_integration_test.go -v

# 또는 직접 연결 테스트 (SQL*Plus가 설치된 경우)
sqlplus username/password@tns_name
```

### 4. 환경변수 요약 (Oracle)

```bash
# ~/.zshrc 또는 ~/.bashrc에 추가
export ORACLE_HOME=/opt/oracle/instantclient_21_x
export TNS_ADMIN=/opt/oracle/wallet
export DYLD_LIBRARY_PATH=$ORACLE_HOME:$DYLD_LIBRARY_PATH  # macOS
export LD_LIBRARY_PATH=$ORACLE_HOME/lib:$LD_LIBRARY_PATH   # Linux

# Oracle 인증 정보 (민감 정보, 별도 관리 권장)
export ORACLE_USERNAME=your_username
export ORACLE_PASSWORD=your_password
```

---

## GCS 환경 설정

### 1. GCP 프로젝트 설정

```bash
# GCP CLI 설치 (아직 없는 경우)
# https://cloud.google.com/sdk/docs/install

# GCP 로그인
gcloud auth login

# 프로젝트 설정
gcloud config set project YOUR_PROJECT_ID

# 필요한 API 활성화
gcloud services enable storage.googleapis.com
```

### 2. 서비스 계정 생성

```bash
# 서비스 계정 생성
gcloud iam service-accounts create oracle-etl-sa \
    --description="Oracle ETL Pipeline Service Account" \
    --display-name="Oracle ETL SA"

# 필요한 권한 부여
gcloud projects add-iam-policy-binding YOUR_PROJECT_ID \
    --member="serviceAccount:oracle-etl-sa@YOUR_PROJECT_ID.iam.gserviceaccount.com" \
    --role="roles/storage.objectAdmin"

# 키 파일 생성
gcloud iam service-accounts keys create /opt/gcp/service-account.json \
    --iam-account=oracle-etl-sa@YOUR_PROJECT_ID.iam.gserviceaccount.com
```

### 3. GCS 버킷 생성

```bash
# 버킷 생성
gsutil mb -l asia-northeast3 gs://oracle-etl-data

# 생명주기 정책 설정 (선택: 90일 후 자동 삭제)
cat > lifecycle.json << EOF
{
  "rule": [
    {
      "action": {"type": "Delete"},
      "condition": {"age": 90}
    }
  ]
}
EOF
gsutil lifecycle set lifecycle.json gs://oracle-etl-data
```

### 4. 환경변수 요약 (GCS)

```bash
# ~/.zshrc 또는 ~/.bashrc에 추가
export GCS_PROJECT_ID=your-gcp-project-id
export GCS_BUCKET_NAME=oracle-etl-data
export GCS_CREDENTIALS_FILE=/opt/gcp/service-account.json

# 또는 Application Default Credentials 사용
gcloud auth application-default login
```

---

## 애플리케이션 설정

### 1. config.yaml 설정

프로젝트 루트의 `config.yaml` 파일을 수정합니다:

```yaml
# config.yaml
server:
  port: 8080
  read_timeout: 30s
  write_timeout: 300s  # ETL 작업을 위한 긴 타임아웃

app:
  name: oracle-etl
  version: 1.0.0
  environment: development  # development | staging | production

# Oracle 연결 설정
oracle:
  wallet_path: /opt/oracle/wallet    # Oracle Wallet 경로
  tns_name: oracle_atp_high          # tnsnames.ora의 연결 이름
  username: ${ORACLE_USERNAME}       # 환경변수 참조
  password: ${ORACLE_PASSWORD}       # 환경변수 참조
  pool_min: 5                        # 최소 커넥션 풀 크기
  pool_max: 20                       # 최대 커넥션 풀 크기
  fetch_array_size: 1000             # 배치 페치 크기

# GCS 설정
gcs:
  project_id: ${GCS_PROJECT_ID}
  bucket_name: oracle-etl-data
  credentials_file: /opt/gcp/service-account.json

# ETL 설정
etl:
  chunk_size: 10000      # 청크당 row 수
  parallel_tables: 4     # 동시 처리 테이블 수
  retry_attempts: 3      # 재시도 횟수
  retry_backoff: 1s      # 재시도 간격

# 인증 설정
auth:
  enabled: true
  api_keys:
    - ${API_KEY}         # 환경변수로 API Key 설정
  bearer_secret: ${JWT_SECRET}
```

### 2. 환경별 설정 파일

```bash
# 개발 환경
cp config.yaml config.development.yaml

# 스테이징 환경
cp config.yaml config.staging.yaml

# 프로덕션 환경
cp config.yaml config.production.yaml

# 환경별 설정 파일 사용
CONFIG_PATH=./config.production.yaml ./bin/oracle-etl
```

### 3. 필수 환경변수 설정

```bash
# ~/.zshrc 또는 ~/.bashrc에 추가

# Oracle 설정
export ORACLE_HOME=/opt/oracle/instantclient_21_x
export TNS_ADMIN=/opt/oracle/wallet
export ORACLE_USERNAME=etl_user
export ORACLE_PASSWORD=secure_password_here

# GCS 설정
export GCS_PROJECT_ID=my-gcp-project
export GCS_BUCKET_NAME=oracle-etl-data
export GCS_CREDENTIALS_FILE=/opt/gcp/service-account.json

# 인증 설정
export API_KEY=your-api-key-here
export JWT_SECRET=your-jwt-secret-here

# 서버 설정 (선택)
export SERVER_PORT=8080
export AUTH_ENABLED=true
```

### 4. .env 파일 사용 (개발용)

```bash
# .env 파일 생성 (gitignore에 추가 필수!)
cat > .env << EOF
ORACLE_USERNAME=etl_user
ORACLE_PASSWORD=dev_password
GCS_PROJECT_ID=my-dev-project
API_KEY=dev-api-key
JWT_SECRET=dev-jwt-secret
EOF

# .gitignore에 추가
echo ".env" >> .gitignore
```

---

## 프론트엔드 설정

### 1. 의존성 설치

```bash
cd web
npm install
```

### 2. 환경변수 설정

```bash
# web/.env.local 파일 생성
cat > .env.local << EOF
# API 서버 URL (개발 시 프록시 사용)
NEXT_PUBLIC_API_URL=http://localhost:8080

# 기타 설정
NEXT_PUBLIC_APP_NAME=Oracle ETL Dashboard
EOF
```

### 3. Next.js 프록시 설정

`web/next.config.ts`에서 API 프록시 설정을 확인합니다:

```typescript
// next.config.ts
const nextConfig = {
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: 'http://localhost:8080/api/:path*',
      },
    ];
  },
};

export default nextConfig;
```

### 4. 개발 서버 실행

```bash
# 백엔드 먼저 실행
cd oracle-etl
go run ./cmd/server

# 다른 터미널에서 프론트엔드 실행
cd oracle-etl/web
npm run dev

# 브라우저에서 http://localhost:3000 접속
```

---

## Docker 환경

### 1. Docker Compose 설정

```yaml
# docker-compose.yml
version: '3.8'

services:
  backend:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      - ORACLE_USERNAME=${ORACLE_USERNAME}
      - ORACLE_PASSWORD=${ORACLE_PASSWORD}
      - GCS_PROJECT_ID=${GCS_PROJECT_ID}
      - API_KEY=${API_KEY}
      - JWT_SECRET=${JWT_SECRET}
    volumes:
      - /opt/oracle/wallet:/opt/wallet:ro
      - /opt/gcp:/opt/gcp:ro
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/api/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  frontend:
    build:
      context: ./web
      dockerfile: Dockerfile
    ports:
      - "3000:3000"
    environment:
      - NEXT_PUBLIC_API_URL=http://backend:8080
    depends_on:
      - backend
```

### 2. 백엔드 Dockerfile

```dockerfile
# Dockerfile
FROM golang:1.23-alpine AS builder

# Oracle Instant Client 설치를 위한 의존성
RUN apk add --no-cache gcc musl-dev libaio-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -o /oracle-etl ./cmd/server

FROM alpine:latest
RUN apk add --no-cache libaio libnsl curl
COPY --from=builder /oracle-etl /oracle-etl
COPY config.yaml /config.yaml

EXPOSE 8080
CMD ["/oracle-etl"]
```

### 3. 프론트엔드 Dockerfile

```dockerfile
# web/Dockerfile
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM node:20-alpine AS runner
WORKDIR /app
ENV NODE_ENV production

COPY --from=builder /app/public ./public
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static

EXPOSE 3000
ENV PORT 3000
CMD ["node", "server.js"]
```

### 4. Docker Compose 실행

```bash
# 환경변수 설정
cp .env.example .env
# .env 파일 편집

# 컨테이너 빌드 및 실행
docker-compose up --build -d

# 로그 확인
docker-compose logs -f

# 중지
docker-compose down
```

---

## 문제 해결

### Oracle 연결 문제

#### ORA-12154: TNS:could not resolve the connect identifier

```bash
# TNS_ADMIN 환경변수 확인
echo $TNS_ADMIN

# tnsnames.ora 파일 존재 확인
ls -la $TNS_ADMIN/tnsnames.ora

# tnsnames.ora에서 연결 이름 확인
cat $TNS_ADMIN/tnsnames.ora | grep -A5 "oracle_atp"
```

#### ORA-28759: failure to open file

```bash
# Wallet 파일 권한 확인
ls -la /opt/oracle/wallet/

# sqlnet.ora 경로 확인
cat $TNS_ADMIN/sqlnet.ora
```

#### godror 빌드 오류

```bash
# Oracle Instant Client 라이브러리 경로 확인
# macOS
export CGO_CFLAGS="-I$ORACLE_HOME/sdk/include"
export CGO_LDFLAGS="-L$ORACLE_HOME"

# Linux
export PKG_CONFIG_PATH=$ORACLE_HOME/lib/pkgconfig

# 다시 빌드
go build ./...
```

### GCS 연결 문제

#### 인증 오류

```bash
# 서비스 계정 키 확인
cat $GCS_CREDENTIALS_FILE | jq .client_email

# 권한 확인
gcloud projects get-iam-policy YOUR_PROJECT_ID \
    --flatten="bindings[].members" \
    --filter="bindings.members:oracle-etl-sa@"
```

#### 버킷 접근 오류

```bash
# 버킷 존재 확인
gsutil ls gs://oracle-etl-data

# 버킷 권한 확인
gsutil iam get gs://oracle-etl-data
```

### 프론트엔드 문제

#### CORS 오류

백엔드에서 CORS 미들웨어가 활성화되어 있는지 확인:

```go
// internal/middleware/cors.go
app.Use(cors.New(cors.Config{
    AllowOrigins: "http://localhost:3000",
    AllowHeaders: "Origin, Content-Type, Accept, X-API-Key",
}))
```

#### API 프록시 오류

```bash
# Next.js 개발 서버 재시작
cd web
npm run dev

# 백엔드 서버 실행 확인
curl http://localhost:8080/api/health
```

---

## 체크리스트

### 개발 환경 설정 완료 체크리스트

- [ ] Go 1.23+ 설치됨
- [ ] Oracle Instant Client 설치됨
- [ ] ORACLE_HOME 환경변수 설정됨
- [ ] Oracle Wallet 다운로드 및 설정됨
- [ ] TNS_ADMIN 환경변수 설정됨
- [ ] GCP 서비스 계정 생성됨
- [ ] GCS 버킷 생성됨
- [ ] GCS_CREDENTIALS_FILE 환경변수 설정됨
- [ ] config.yaml 설정됨
- [ ] Node.js 20+ 설치됨
- [ ] 프론트엔드 의존성 설치됨 (npm install)
- [ ] 백엔드 서버 실행 확인됨
- [ ] 프론트엔드 서버 실행 확인됨
- [ ] API 연결 테스트 완료됨

---

## 추가 리소스

- [Oracle Instant Client 공식 문서](https://www.oracle.com/database/technologies/instant-client.html)
- [godror 드라이버 문서](https://github.com/godror/godror)
- [GCS Go 클라이언트 문서](https://pkg.go.dev/cloud.google.com/go/storage)
- [Next.js 16 공식 문서](https://nextjs.org/docs)
