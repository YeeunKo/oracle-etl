# GCP Platform Examples

## Go SDK Examples

### Cloud Storage Client Setup

```go
package main

import (
    "context"
    "fmt"
    "io"
    "log"
    "os"
    "time"

    "cloud.google.com/go/storage"
    "google.golang.org/api/iterator"
)

// StorageClient wraps the GCS client with common operations
type StorageClient struct {
    client *storage.Client
    bucket string
}

// NewStorageClient creates a new storage client
func NewStorageClient(ctx context.Context, bucketName string) (*StorageClient, error) {
    client, err := storage.NewClient(ctx)
    if err != nil {
        return nil, fmt.Errorf("storage.NewClient: %w", err)
    }
    return &StorageClient{
        client: client,
        bucket: bucketName,
    }, nil
}

// Close closes the storage client
func (s *StorageClient) Close() error {
    return s.client.Close()
}
```

### Upload Object

```go
// UploadObject uploads data to GCS
func (s *StorageClient) UploadObject(ctx context.Context, objectName string, data []byte, contentType string) error {
    bucket := s.client.Bucket(s.bucket)
    obj := bucket.Object(objectName)

    // 작성자 생성
    wc := obj.NewWriter(ctx)
    wc.ContentType = contentType
    wc.Metadata = map[string]string{
        "uploaded-at": time.Now().Format(time.RFC3339),
    }

    // 데이터 쓰기
    if _, err := wc.Write(data); err != nil {
        return fmt.Errorf("Writer.Write: %w", err)
    }

    // 작성자 닫기 (필수 - 업로드 완료)
    if err := wc.Close(); err != nil {
        return fmt.Errorf("Writer.Close: %w", err)
    }

    log.Printf("Uploaded %s to gs://%s/%s", objectName, s.bucket, objectName)
    return nil
}

// UploadFile uploads a file from local filesystem
func (s *StorageClient) UploadFile(ctx context.Context, objectName, filePath string) error {
    f, err := os.Open(filePath)
    if err != nil {
        return fmt.Errorf("os.Open: %w", err)
    }
    defer f.Close()

    bucket := s.client.Bucket(s.bucket)
    obj := bucket.Object(objectName)
    wc := obj.NewWriter(ctx)

    if _, err := io.Copy(wc, f); err != nil {
        return fmt.Errorf("io.Copy: %w", err)
    }

    if err := wc.Close(); err != nil {
        return fmt.Errorf("Writer.Close: %w", err)
    }

    return nil
}
```

### Streaming Upload

```go
// StreamingUpload uploads data using streaming for large files
func (s *StorageClient) StreamingUpload(ctx context.Context, objectName string, reader io.Reader, contentType string) error {
    bucket := s.client.Bucket(s.bucket)
    obj := bucket.Object(objectName)

    wc := obj.NewWriter(ctx)
    wc.ContentType = contentType
    wc.ChunkSize = 16 * 1024 * 1024 // 16MB chunks

    // 스트리밍 복사
    written, err := io.Copy(wc, reader)
    if err != nil {
        wc.Close() // 에러 시에도 닫기 시도
        return fmt.Errorf("io.Copy: %w", err)
    }

    if err := wc.Close(); err != nil {
        return fmt.Errorf("Writer.Close: %w", err)
    }

    log.Printf("Streamed %d bytes to gs://%s/%s", written, s.bucket, objectName)
    return nil
}
```

### Download Object

```go
// DownloadObject downloads an object from GCS
func (s *StorageClient) DownloadObject(ctx context.Context, objectName string) ([]byte, error) {
    bucket := s.client.Bucket(s.bucket)
    obj := bucket.Object(objectName)

    rc, err := obj.NewReader(ctx)
    if err != nil {
        return nil, fmt.Errorf("Object.NewReader: %w", err)
    }
    defer rc.Close()

    data, err := io.ReadAll(rc)
    if err != nil {
        return nil, fmt.Errorf("io.ReadAll: %w", err)
    }

    return data, nil
}

// DownloadToFile downloads an object to local filesystem
func (s *StorageClient) DownloadToFile(ctx context.Context, objectName, filePath string) error {
    bucket := s.client.Bucket(s.bucket)
    obj := bucket.Object(objectName)

    rc, err := obj.NewReader(ctx)
    if err != nil {
        return fmt.Errorf("Object.NewReader: %w", err)
    }
    defer rc.Close()

    f, err := os.Create(filePath)
    if err != nil {
        return fmt.Errorf("os.Create: %w", err)
    }
    defer f.Close()

    if _, err := io.Copy(f, rc); err != nil {
        return fmt.Errorf("io.Copy: %w", err)
    }

    return nil
}
```

### List Objects

```go
// ListObjects lists objects in a bucket with optional prefix
func (s *StorageClient) ListObjects(ctx context.Context, prefix string) ([]string, error) {
    bucket := s.client.Bucket(s.bucket)

    var objects []string
    it := bucket.Objects(ctx, &storage.Query{
        Prefix: prefix,
    })

    for {
        attrs, err := it.Next()
        if err == iterator.Done {
            break
        }
        if err != nil {
            return nil, fmt.Errorf("Bucket.Objects: %w", err)
        }
        objects = append(objects, attrs.Name)
    }

    return objects, nil
}

// ListObjectsWithMetadata lists objects with their metadata
func (s *StorageClient) ListObjectsWithMetadata(ctx context.Context, prefix string) ([]*storage.ObjectAttrs, error) {
    bucket := s.client.Bucket(s.bucket)

    var objects []*storage.ObjectAttrs
    it := bucket.Objects(ctx, &storage.Query{
        Prefix: prefix,
    })

    for {
        attrs, err := it.Next()
        if err == iterator.Done {
            break
        }
        if err != nil {
            return nil, fmt.Errorf("Bucket.Objects: %w", err)
        }
        objects = append(objects, attrs)
    }

    return objects, nil
}
```

### Delete Object

```go
// DeleteObject deletes an object from GCS
func (s *StorageClient) DeleteObject(ctx context.Context, objectName string) error {
    bucket := s.client.Bucket(s.bucket)
    obj := bucket.Object(objectName)

    if err := obj.Delete(ctx); err != nil {
        return fmt.Errorf("Object.Delete: %w", err)
    }

    log.Printf("Deleted gs://%s/%s", s.bucket, objectName)
    return nil
}

// DeleteObjects deletes multiple objects
func (s *StorageClient) DeleteObjects(ctx context.Context, objectNames []string) error {
    for _, name := range objectNames {
        if err := s.DeleteObject(ctx, name); err != nil {
            return err
        }
    }
    return nil
}
```

### Signed URL Generation

```go
// GenerateSignedURL generates a signed URL for temporary access
func (s *StorageClient) GenerateSignedURL(ctx context.Context, objectName string, expiration time.Duration, method string) (string, error) {
    bucket := s.client.Bucket(s.bucket)

    opts := &storage.SignedURLOptions{
        Method:  method, // "GET", "PUT", etc.
        Expires: time.Now().Add(expiration),
    }

    url, err := bucket.SignedURL(objectName, opts)
    if err != nil {
        return "", fmt.Errorf("Bucket.SignedURL: %w", err)
    }

    return url, nil
}
```

### Complete ETL Example

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "time"
)

// ETLRecord represents a data record
type ETLRecord struct {
    ID        string    `json:"id"`
    Data      string    `json:"data"`
    Timestamp time.Time `json:"timestamp"`
}

func main() {
    ctx := context.Background()

    // 클라이언트 생성
    client, err := NewStorageClient(ctx, "my-etl-bucket")
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    defer client.Close()

    // 데이터 생성
    records := []ETLRecord{
        {ID: "1", Data: "sample data 1", Timestamp: time.Now()},
        {ID: "2", Data: "sample data 2", Timestamp: time.Now()},
    }

    // JSON으로 직렬화
    data, err := json.Marshal(records)
    if err != nil {
        log.Fatalf("Failed to marshal: %v", err)
    }

    // 업로드
    objectName := fmt.Sprintf("etl-output/%s/data.json", time.Now().Format("2006-01-02"))
    if err := client.UploadObject(ctx, objectName, data, "application/json"); err != nil {
        log.Fatalf("Failed to upload: %v", err)
    }

    // 확인을 위해 다시 다운로드
    downloaded, err := client.DownloadObject(ctx, objectName)
    if err != nil {
        log.Fatalf("Failed to download: %v", err)
    }

    log.Printf("Successfully uploaded and verified: %s", string(downloaded))
}
```

---

## Terraform Complete Examples

### ETL Infrastructure Module

```hcl
# modules/etl-infrastructure/main.tf

terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

variable "project_id" {
  description = "GCP 프로젝트 ID"
  type        = string
}

variable "region" {
  description = "GCP 리전"
  type        = string
  default     = "asia-northeast3"
}

variable "environment" {
  description = "환경 (dev, staging, prod)"
  type        = string
  default     = "dev"
}

# 서비스 계정
resource "google_service_account" "etl" {
  account_id   = "etl-${var.environment}"
  display_name = "ETL Service Account (${var.environment})"
  description  = "Service account for ETL pipeline in ${var.environment}"
}

# 입력 버킷
resource "google_storage_bucket" "input" {
  name     = "${var.project_id}-etl-input-${var.environment}"
  location = var.region

  uniform_bucket_level_access = true

  lifecycle_rule {
    condition {
      age = 7
    }
    action {
      type = "Delete"
    }
  }

  labels = {
    environment = var.environment
    purpose     = "etl-input"
  }
}

# 출력 버킷
resource "google_storage_bucket" "output" {
  name     = "${var.project_id}-etl-output-${var.environment}"
  location = var.region

  uniform_bucket_level_access = true

  versioning {
    enabled = true
  }

  lifecycle_rule {
    condition {
      age = 30
    }
    action {
      type          = "SetStorageClass"
      storage_class = "NEARLINE"
    }
  }

  lifecycle_rule {
    condition {
      age = 90
    }
    action {
      type          = "SetStorageClass"
      storage_class = "COLDLINE"
    }
  }

  labels = {
    environment = var.environment
    purpose     = "etl-output"
  }
}

# IAM 바인딩
resource "google_storage_bucket_iam_member" "input_reader" {
  bucket = google_storage_bucket.input.name
  role   = "roles/storage.objectViewer"
  member = "serviceAccount:${google_service_account.etl.email}"
}

resource "google_storage_bucket_iam_member" "output_writer" {
  bucket = google_storage_bucket.output.name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${google_service_account.etl.email}"
}

# 출력
output "service_account_email" {
  value = google_service_account.etl.email
}

output "input_bucket" {
  value = google_storage_bucket.input.name
}

output "output_bucket" {
  value = google_storage_bucket.output.name
}
```

### Cloud Run API Service

```hcl
# cloud-run-api/main.tf

resource "google_artifact_registry_repository" "api" {
  location      = var.region
  repository_id = "api-${var.environment}"
  format        = "DOCKER"
  description   = "Docker repository for API service"
}

resource "google_service_account" "cloudrun" {
  account_id   = "cloudrun-api-${var.environment}"
  display_name = "Cloud Run API Service Account"
}

resource "google_cloud_run_v2_service" "api" {
  name     = "api-${var.environment}"
  location = var.region

  template {
    containers {
      image = "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.api.repository_id}/api:latest"

      ports {
        container_port = 8080
      }

      env {
        name  = "ENV"
        value = var.environment
      }

      env {
        name  = "GCS_BUCKET"
        value = google_storage_bucket.output.name
      }

      resources {
        limits = {
          cpu    = var.environment == "prod" ? "2000m" : "1000m"
          memory = var.environment == "prod" ? "1Gi" : "512Mi"
        }
      }

      startup_probe {
        http_get {
          path = "/health"
          port = 8080
        }
        initial_delay_seconds = 5
        timeout_seconds       = 3
        period_seconds        = 10
        failure_threshold     = 3
      }

      liveness_probe {
        http_get {
          path = "/health"
          port = 8080
        }
        period_seconds    = 30
        timeout_seconds   = 5
        failure_threshold = 3
      }
    }

    scaling {
      min_instance_count = var.environment == "prod" ? 1 : 0
      max_instance_count = var.environment == "prod" ? 10 : 3
    }

    service_account = google_service_account.cloudrun.email
  }

  traffic {
    type    = "TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST"
    percent = 100
  }

  depends_on = [
    google_artifact_registry_repository.api
  ]
}

# 공개 액세스 (필요시)
resource "google_cloud_run_service_iam_member" "invoker" {
  count    = var.public_access ? 1 : 0
  location = google_cloud_run_v2_service.api.location
  service  = google_cloud_run_v2_service.api.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}

# Cloud Storage 접근 권한
resource "google_storage_bucket_iam_member" "cloudrun_storage" {
  bucket = google_storage_bucket.output.name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${google_service_account.cloudrun.email}"
}

output "service_url" {
  value = google_cloud_run_v2_service.api.uri
}
```

### Compute Engine Worker

```hcl
# compute-worker/main.tf

resource "google_compute_instance_template" "etl_worker" {
  name_prefix  = "etl-worker-"
  machine_type = var.environment == "prod" ? "e2-standard-4" : "e2-medium"
  region       = var.region

  disk {
    source_image = "ubuntu-os-cloud/ubuntu-2204-lts"
    disk_size_gb = 50
    disk_type    = "pd-ssd"
    auto_delete  = true
    boot         = true
  }

  network_interface {
    network = "default"
    access_config {
      // 외부 IP
    }
  }

  service_account {
    email  = google_service_account.etl.email
    scopes = ["cloud-platform"]
  }

  metadata_startup_script = <<-EOF
    #!/bin/bash
    set -e

    # 로깅 설정
    exec > >(tee /var/log/startup-script.log|logger -t startup-script -s 2>/dev/console) 2>&1

    # 패키지 업데이트
    apt-get update
    apt-get install -y docker.io docker-compose

    # Docker 시작
    systemctl start docker
    systemctl enable docker

    # ETL 컨테이너 실행
    docker pull gcr.io/${var.project_id}/etl-worker:latest
    docker run -d \
      --restart=always \
      --name=etl-worker \
      -e GCS_INPUT_BUCKET=${google_storage_bucket.input.name} \
      -e GCS_OUTPUT_BUCKET=${google_storage_bucket.output.name} \
      gcr.io/${var.project_id}/etl-worker:latest

    echo "Startup script completed"
  EOF

  labels = {
    environment = var.environment
    role        = "etl-worker"
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "google_compute_instance_group_manager" "etl_workers" {
  name               = "etl-workers-${var.environment}"
  base_instance_name = "etl-worker"
  zone               = "${var.region}-a"
  target_size        = var.environment == "prod" ? 2 : 1

  version {
    instance_template = google_compute_instance_template.etl_worker.id
  }

  named_port {
    name = "http"
    port = 8080
  }

  auto_healing_policies {
    health_check      = google_compute_health_check.etl.id
    initial_delay_sec = 300
  }
}

resource "google_compute_health_check" "etl" {
  name               = "etl-health-check-${var.environment}"
  check_interval_sec = 30
  timeout_sec        = 10

  http_health_check {
    port         = 8080
    request_path = "/health"
  }
}
```

---

## GitHub Actions CI/CD

### Deploy to Cloud Run

```yaml
# .github/workflows/deploy-cloudrun.yml
name: Deploy to Cloud Run

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

env:
  PROJECT_ID: ${{ secrets.GCP_PROJECT_ID }}
  REGION: asia-northeast3
  SERVICE_NAME: api-service

jobs:
  build-and-deploy:
    runs-on: ubuntu-latest

    permissions:
      contents: read
      id-token: write

    steps:
      - uses: actions/checkout@v4

      - id: auth
        uses: google-github-actions/auth@v2
        with:
          workload_identity_provider: ${{ secrets.WIF_PROVIDER }}
          service_account: ${{ secrets.WIF_SERVICE_ACCOUNT }}

      - uses: google-github-actions/setup-gcloud@v2
        with:
          project_id: ${{ env.PROJECT_ID }}

      - name: Configure Docker
        run: gcloud auth configure-docker ${{ env.REGION }}-docker.pkg.dev

      - name: Build and Push
        run: |
          docker build -t ${{ env.REGION }}-docker.pkg.dev/${{ env.PROJECT_ID }}/api/${{ env.SERVICE_NAME }}:${{ github.sha }} .
          docker push ${{ env.REGION }}-docker.pkg.dev/${{ env.PROJECT_ID }}/api/${{ env.SERVICE_NAME }}:${{ github.sha }}

      - name: Deploy to Cloud Run
        if: github.ref == 'refs/heads/main'
        run: |
          gcloud run deploy ${{ env.SERVICE_NAME }} \
            --image=${{ env.REGION }}-docker.pkg.dev/${{ env.PROJECT_ID }}/api/${{ env.SERVICE_NAME }}:${{ github.sha }} \
            --region=${{ env.REGION }} \
            --platform=managed \
            --allow-unauthenticated
```

### Terraform Apply

```yaml
# .github/workflows/terraform.yml
name: Terraform

on:
  push:
    branches: [main]
    paths: ['terraform/**']
  pull_request:
    branches: [main]
    paths: ['terraform/**']

jobs:
  terraform:
    runs-on: ubuntu-latest

    permissions:
      contents: read
      id-token: write
      pull-requests: write

    defaults:
      run:
        working-directory: terraform

    steps:
      - uses: actions/checkout@v4

      - id: auth
        uses: google-github-actions/auth@v2
        with:
          workload_identity_provider: ${{ secrets.WIF_PROVIDER }}
          service_account: ${{ secrets.WIF_SERVICE_ACCOUNT }}

      - uses: hashicorp/setup-terraform@v3

      - name: Terraform Init
        run: terraform init

      - name: Terraform Format
        run: terraform fmt -check

      - name: Terraform Plan
        run: terraform plan -no-color
        continue-on-error: true

      - name: Terraform Apply
        if: github.ref == 'refs/heads/main' && github.event_name == 'push'
        run: terraform apply -auto-approve
```
