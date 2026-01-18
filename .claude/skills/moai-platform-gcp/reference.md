# GCP Platform Reference

## gcloud CLI Commands

### Authentication

```bash
# 사용자 계정으로 로그인
gcloud auth login

# 애플리케이션 기본 인증 설정
gcloud auth application-default login

# 현재 인증 정보 확인
gcloud auth list

# 프로젝트 설정
gcloud config set project PROJECT_ID

# 리전 설정
gcloud config set compute/region asia-northeast3

# 존 설정
gcloud config set compute/zone asia-northeast3-a
```

### Cloud Storage (gsutil)

```bash
# 버킷 생성
gsutil mb -l ASIA-NORTHEAST3 gs://bucket-name

# 버킷에 균일 버킷 수준 액세스 활성화
gsutil uniformbucketlevelaccess set on gs://bucket-name

# 파일 업로드
gsutil cp local-file.txt gs://bucket-name/

# 디렉토리 업로드 (재귀)
gsutil -m cp -r local-dir/ gs://bucket-name/

# 파일 다운로드
gsutil cp gs://bucket-name/file.txt ./

# 버킷 내용 나열
gsutil ls gs://bucket-name/

# 상세 정보 포함 나열
gsutil ls -l gs://bucket-name/

# 파일 삭제
gsutil rm gs://bucket-name/file.txt

# 디렉토리 삭제 (재귀)
gsutil rm -r gs://bucket-name/path/

# 버킷 삭제
gsutil rb gs://bucket-name

# 수명 주기 정책 설정
gsutil lifecycle set lifecycle.json gs://bucket-name

# 메타데이터 설정
gsutil setmeta -h "Content-Type:application/json" gs://bucket-name/file.json

# CORS 설정
gsutil cors set cors.json gs://bucket-name

# 서명된 URL 생성 (1시간 유효)
gsutil signurl -d 1h service-account.json gs://bucket-name/file.txt
```

### Compute Engine

```bash
# 인스턴스 생성
gcloud compute instances create INSTANCE_NAME \
    --machine-type=e2-standard-4 \
    --zone=asia-northeast3-a \
    --image-family=ubuntu-2204-lts \
    --image-project=ubuntu-os-cloud \
    --boot-disk-size=50GB \
    --boot-disk-type=pd-ssd

# 시작 스크립트 포함 인스턴스 생성
gcloud compute instances create INSTANCE_NAME \
    --machine-type=e2-medium \
    --zone=asia-northeast3-a \
    --image-family=ubuntu-2204-lts \
    --image-project=ubuntu-os-cloud \
    --metadata-from-file=startup-script=startup.sh

# 인스턴스 목록
gcloud compute instances list

# 인스턴스 중지
gcloud compute instances stop INSTANCE_NAME --zone=asia-northeast3-a

# 인스턴스 시작
gcloud compute instances start INSTANCE_NAME --zone=asia-northeast3-a

# 인스턴스 삭제
gcloud compute instances delete INSTANCE_NAME --zone=asia-northeast3-a

# SSH 접속
gcloud compute ssh INSTANCE_NAME --zone=asia-northeast3-a

# 파일 전송 (SCP)
gcloud compute scp local-file.txt INSTANCE_NAME:~/

# 방화벽 규칙 생성
gcloud compute firewall-rules create allow-http \
    --allow=tcp:80 \
    --source-ranges=0.0.0.0/0 \
    --target-tags=http-server

# 머신 타입 목록
gcloud compute machine-types list --filter="zone:asia-northeast3-a"

# 이미지 목록
gcloud compute images list --project=ubuntu-os-cloud
```

### Cloud Run

```bash
# 소스에서 배포
gcloud run deploy SERVICE_NAME \
    --source=. \
    --region=asia-northeast3 \
    --allow-unauthenticated

# 컨테이너 이미지로 배포
gcloud run deploy SERVICE_NAME \
    --image=gcr.io/PROJECT_ID/IMAGE_NAME \
    --region=asia-northeast3 \
    --platform=managed

# 환경 변수 설정
gcloud run deploy SERVICE_NAME \
    --image=gcr.io/PROJECT_ID/IMAGE_NAME \
    --region=asia-northeast3 \
    --set-env-vars="KEY1=value1,KEY2=value2"

# 메모리 및 CPU 설정
gcloud run deploy SERVICE_NAME \
    --image=gcr.io/PROJECT_ID/IMAGE_NAME \
    --region=asia-northeast3 \
    --memory=512Mi \
    --cpu=1

# 최소/최대 인스턴스 설정
gcloud run deploy SERVICE_NAME \
    --image=gcr.io/PROJECT_ID/IMAGE_NAME \
    --region=asia-northeast3 \
    --min-instances=1 \
    --max-instances=10

# 서비스 목록
gcloud run services list --region=asia-northeast3

# 서비스 상세 정보
gcloud run services describe SERVICE_NAME --region=asia-northeast3

# 서비스 삭제
gcloud run services delete SERVICE_NAME --region=asia-northeast3

# 커스텀 도메인 매핑
gcloud run domain-mappings create \
    --service=SERVICE_NAME \
    --domain=example.com \
    --region=asia-northeast3

# 트래픽 분할
gcloud run services update-traffic SERVICE_NAME \
    --to-revisions=REVISION1=50,REVISION2=50 \
    --region=asia-northeast3
```

### IAM & Service Accounts

```bash
# 서비스 계정 생성
gcloud iam service-accounts create SA_NAME \
    --display-name="Service Account Display Name"

# 서비스 계정 목록
gcloud iam service-accounts list

# 역할 부여
gcloud projects add-iam-policy-binding PROJECT_ID \
    --member="serviceAccount:SA_NAME@PROJECT_ID.iam.gserviceaccount.com" \
    --role="roles/storage.admin"

# 서비스 계정 키 생성
gcloud iam service-accounts keys create key.json \
    --iam-account=SA_NAME@PROJECT_ID.iam.gserviceaccount.com

# 서비스 계정 키 목록
gcloud iam service-accounts keys list \
    --iam-account=SA_NAME@PROJECT_ID.iam.gserviceaccount.com

# IAM 정책 조회
gcloud projects get-iam-policy PROJECT_ID

# 사용 가능한 역할 목록
gcloud iam roles list

# 커스텀 역할 생성
gcloud iam roles create ROLE_ID \
    --project=PROJECT_ID \
    --title="Custom Role Title" \
    --permissions="storage.buckets.get,storage.buckets.list"
```

---

## Terraform Configuration

### Provider Configuration

```hcl
terraform {
  required_version = ">= 1.0.0"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
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
```

### Cloud Storage Bucket

```hcl
resource "google_storage_bucket" "data_bucket" {
  name     = "${var.project_id}-data-bucket"
  location = var.region

  # 균일 버킷 수준 액세스 (권장)
  uniform_bucket_level_access = true

  # 버전 관리
  versioning {
    enabled = true
  }

  # 수명 주기 정책
  lifecycle_rule {
    condition {
      age = 30  # 30일 후
    }
    action {
      type          = "SetStorageClass"
      storage_class = "NEARLINE"
    }
  }

  lifecycle_rule {
    condition {
      age = 90  # 90일 후
    }
    action {
      type          = "SetStorageClass"
      storage_class = "COLDLINE"
    }
  }

  lifecycle_rule {
    condition {
      age = 365  # 365일 후 삭제
    }
    action {
      type = "Delete"
    }
  }

  # 삭제 방지 (선택)
  force_destroy = false

  labels = {
    environment = "production"
    purpose     = "etl-data"
  }
}

# 버킷 IAM
resource "google_storage_bucket_iam_member" "viewer" {
  bucket = google_storage_bucket.data_bucket.name
  role   = "roles/storage.objectViewer"
  member = "serviceAccount:${google_service_account.etl_sa.email}"
}
```

### Compute Engine Instance

```hcl
resource "google_compute_instance" "etl_worker" {
  name         = "etl-worker"
  machine_type = "e2-standard-4"
  zone         = "${var.region}-a"

  boot_disk {
    initialize_params {
      image = "ubuntu-os-cloud/ubuntu-2204-lts"
      size  = 50
      type  = "pd-ssd"
    }
  }

  network_interface {
    network = "default"
    access_config {
      # 외부 IP 할당
    }
  }

  # 서비스 계정
  service_account {
    email  = google_service_account.etl_sa.email
    scopes = ["cloud-platform"]
  }

  # 시작 스크립트
  metadata_startup_script = <<-EOF
    #!/bin/bash
    apt-get update
    apt-get install -y docker.io
    systemctl start docker
    systemctl enable docker
  EOF

  tags = ["http-server", "https-server"]

  labels = {
    environment = "production"
    purpose     = "etl"
  }
}

# 방화벽 규칙
resource "google_compute_firewall" "allow_http" {
  name    = "allow-http"
  network = "default"

  allow {
    protocol = "tcp"
    ports    = ["80", "443"]
  }

  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["http-server", "https-server"]
}
```

### Cloud Run Service

```hcl
resource "google_cloud_run_v2_service" "api" {
  name     = "api-service"
  location = var.region

  template {
    containers {
      image = "gcr.io/${var.project_id}/api:latest"

      ports {
        container_port = 8080
      }

      env {
        name  = "ENV"
        value = "production"
      }

      env {
        name = "DATABASE_URL"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.db_url.secret_id
            version = "latest"
          }
        }
      }

      resources {
        limits = {
          cpu    = "1000m"
          memory = "512Mi"
        }
      }
    }

    scaling {
      min_instance_count = 0
      max_instance_count = 10
    }

    service_account = google_service_account.cloudrun_sa.email
  }

  traffic {
    type    = "TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST"
    percent = 100
  }
}

# 공개 액세스 허용
resource "google_cloud_run_service_iam_member" "public" {
  location = google_cloud_run_v2_service.api.location
  service  = google_cloud_run_v2_service.api.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}
```

### Service Account

```hcl
resource "google_service_account" "etl_sa" {
  account_id   = "etl-service-account"
  display_name = "ETL Service Account"
  description  = "Service account for ETL pipeline"
}

resource "google_project_iam_member" "etl_storage_admin" {
  project = var.project_id
  role    = "roles/storage.admin"
  member  = "serviceAccount:${google_service_account.etl_sa.email}"
}

resource "google_project_iam_member" "etl_compute_viewer" {
  project = var.project_id
  role    = "roles/compute.viewer"
  member  = "serviceAccount:${google_service_account.etl_sa.email}"
}

# 서비스 계정 키 (권장하지 않음 - Workload Identity 사용 권장)
resource "google_service_account_key" "etl_sa_key" {
  service_account_id = google_service_account.etl_sa.name
}
```

---

## Storage Classes

| 클래스 | 최소 저장 기간 | 사용 사례 |
|--------|---------------|----------|
| STANDARD | 없음 | 자주 액세스되는 데이터 |
| NEARLINE | 30일 | 월 1회 미만 액세스 |
| COLDLINE | 90일 | 분기 1회 미만 액세스 |
| ARCHIVE | 365일 | 연 1회 미만 액세스 |

---

## Machine Types

### General Purpose (E2)

| 타입 | vCPU | 메모리 |
|------|------|--------|
| e2-micro | 0.25 | 1GB |
| e2-small | 0.5 | 2GB |
| e2-medium | 1 | 4GB |
| e2-standard-2 | 2 | 8GB |
| e2-standard-4 | 4 | 16GB |
| e2-standard-8 | 8 | 32GB |

### Compute Optimized (C2)

| 타입 | vCPU | 메모리 |
|------|------|--------|
| c2-standard-4 | 4 | 16GB |
| c2-standard-8 | 8 | 32GB |
| c2-standard-16 | 16 | 64GB |

### Memory Optimized (M2)

| 타입 | vCPU | 메모리 |
|------|------|--------|
| m2-ultramem-208 | 208 | 5.75TB |
| m2-megamem-416 | 416 | 5.75TB |

---

## Cloud Run Limits

| 리소스 | 1세대 | 2세대 |
|--------|-------|-------|
| 최대 메모리 | 8GB | 32GB |
| 최대 CPU | 4 | 8 |
| 요청 타임아웃 | 60분 | 60분 |
| 최대 인스턴스 | 1000 | 1000 |
| 컨테이너 동시성 | 1000 | 1000 |

---

## IAM Predefined Roles

### Storage

- `roles/storage.admin` - 전체 관리
- `roles/storage.objectAdmin` - 객체 관리
- `roles/storage.objectViewer` - 객체 읽기
- `roles/storage.objectCreator` - 객체 생성

### Compute

- `roles/compute.admin` - 전체 관리
- `roles/compute.instanceAdmin` - 인스턴스 관리
- `roles/compute.viewer` - 읽기 전용

### Cloud Run

- `roles/run.admin` - 전체 관리
- `roles/run.developer` - 배포 권한
- `roles/run.invoker` - 서비스 호출
- `roles/run.viewer` - 읽기 전용
