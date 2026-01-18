# IAM Module

## Overview

Google Cloud IAM(Identity and Access Management)은 리소스 접근 제어를 관리합니다. 서비스 계정, 역할 바인딩, Workload Identity를 통해 보안을 강화합니다.

---

## Service Accounts

### Creating Service Accounts

```hcl
resource "google_service_account" "etl" {
  account_id   = "etl-service-account"
  display_name = "ETL Service Account"
  description  = "Service account for ETL pipeline operations"
  project      = var.project_id
}
```

### gcloud 명령어

```bash
# 서비스 계정 생성
gcloud iam service-accounts create etl-sa \
    --display-name="ETL Service Account" \
    --description="Service account for ETL pipeline"

# 서비스 계정 목록
gcloud iam service-accounts list

# 서비스 계정 상세 정보
gcloud iam service-accounts describe etl-sa@project-id.iam.gserviceaccount.com
```

---

## Role Bindings

### Project-Level Binding

```hcl
resource "google_project_iam_member" "storage_admin" {
  project = var.project_id
  role    = "roles/storage.admin"
  member  = "serviceAccount:${google_service_account.etl.email}"
}

resource "google_project_iam_member" "bigquery_editor" {
  project = var.project_id
  role    = "roles/bigquery.dataEditor"
  member  = "serviceAccount:${google_service_account.etl.email}"
}
```

### Resource-Level Binding

특정 리소스에만 권한을 부여합니다.

```hcl
# 특정 버킷에만 권한 부여
resource "google_storage_bucket_iam_member" "bucket_viewer" {
  bucket = google_storage_bucket.data.name
  role   = "roles/storage.objectViewer"
  member = "serviceAccount:${google_service_account.etl.email}"
}

# 특정 Cloud Run 서비스 호출 권한
resource "google_cloud_run_service_iam_member" "invoker" {
  location = google_cloud_run_v2_service.api.location
  service  = google_cloud_run_v2_service.api.name
  role     = "roles/run.invoker"
  member   = "serviceAccount:${google_service_account.caller.email}"
}
```

### gcloud 명령어

```bash
# 프로젝트 수준 역할 부여
gcloud projects add-iam-policy-binding PROJECT_ID \
    --member="serviceAccount:etl-sa@PROJECT_ID.iam.gserviceaccount.com" \
    --role="roles/storage.admin"

# 버킷 수준 역할 부여
gsutil iam ch \
    serviceAccount:etl-sa@PROJECT_ID.iam.gserviceaccount.com:objectViewer \
    gs://bucket-name

# 역할 제거
gcloud projects remove-iam-policy-binding PROJECT_ID \
    --member="serviceAccount:etl-sa@PROJECT_ID.iam.gserviceaccount.com" \
    --role="roles/storage.admin"
```

---

## Custom Roles

사전 정의된 역할이 너무 많은 권한을 부여할 때 커스텀 역할을 생성합니다.

```hcl
resource "google_project_iam_custom_role" "etl_worker" {
  role_id     = "etlWorker"
  title       = "ETL Worker"
  description = "Custom role for ETL worker with minimal permissions"
  permissions = [
    "storage.objects.get",
    "storage.objects.list",
    "storage.objects.create",
    "storage.buckets.get",
    "bigquery.tables.getData",
    "bigquery.tables.updateData",
    "bigquery.jobs.create",
  ]
}

resource "google_project_iam_member" "etl_custom" {
  project = var.project_id
  role    = google_project_iam_custom_role.etl_worker.id
  member  = "serviceAccount:${google_service_account.etl.email}"
}
```

### gcloud 명령어

```bash
# 커스텀 역할 생성
gcloud iam roles create etlWorker \
    --project=PROJECT_ID \
    --title="ETL Worker" \
    --description="Custom role for ETL worker" \
    --permissions="storage.objects.get,storage.objects.list,storage.objects.create"

# 커스텀 역할 업데이트
gcloud iam roles update etlWorker \
    --project=PROJECT_ID \
    --add-permissions="bigquery.jobs.create"

# 커스텀 역할 목록
gcloud iam roles list --project=PROJECT_ID
```

---

## Service Account Keys

주의: 서비스 계정 키는 보안 위험이 있으므로 Workload Identity 사용을 권장합니다.

### Key 생성 (권장하지 않음)

```hcl
resource "google_service_account_key" "etl_key" {
  service_account_id = google_service_account.etl.name
  public_key_type    = "TYPE_X509_PEM_FILE"
}

# 키를 Secret Manager에 저장
resource "google_secret_manager_secret" "sa_key" {
  secret_id = "etl-sa-key"

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "sa_key" {
  secret      = google_secret_manager_secret.sa_key.id
  secret_data = base64decode(google_service_account_key.etl_key.private_key)
}
```

### gcloud 명령어

```bash
# 키 생성
gcloud iam service-accounts keys create key.json \
    --iam-account=etl-sa@PROJECT_ID.iam.gserviceaccount.com

# 키 목록
gcloud iam service-accounts keys list \
    --iam-account=etl-sa@PROJECT_ID.iam.gserviceaccount.com

# 키 삭제
gcloud iam service-accounts keys delete KEY_ID \
    --iam-account=etl-sa@PROJECT_ID.iam.gserviceaccount.com
```

---

## Workload Identity

GKE 또는 외부 워크로드에서 서비스 계정 키 없이 인증합니다.

### GKE Workload Identity

```hcl
# GKE 서비스 계정
resource "google_service_account" "gke_workload" {
  account_id   = "gke-workload-sa"
  display_name = "GKE Workload Identity SA"
}

# Kubernetes 서비스 계정에 바인딩
resource "google_service_account_iam_member" "workload_identity" {
  service_account_id = google_service_account.gke_workload.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "serviceAccount:${var.project_id}.svc.id.goog[${var.namespace}/${var.k8s_sa_name}]"
}

# Cloud Storage 접근 권한
resource "google_project_iam_member" "gke_storage" {
  project = var.project_id
  role    = "roles/storage.objectAdmin"
  member  = "serviceAccount:${google_service_account.gke_workload.email}"
}
```

### Kubernetes 설정

```yaml
# Kubernetes ServiceAccount
apiVersion: v1
kind: ServiceAccount
metadata:
  name: etl-workload
  namespace: default
  annotations:
    iam.gke.io/gcp-service-account: gke-workload-sa@PROJECT_ID.iam.gserviceaccount.com
```

---

## Workload Identity Federation

외부 ID 공급자(GitHub Actions, AWS 등)에서 GCP 리소스에 접근합니다.

### GitHub Actions 설정

```hcl
# Workload Identity Pool
resource "google_iam_workload_identity_pool" "github" {
  workload_identity_pool_id = "github-pool"
  display_name              = "GitHub Actions Pool"
  description               = "Pool for GitHub Actions workflows"
}

# Provider
resource "google_iam_workload_identity_pool_provider" "github" {
  workload_identity_pool_id          = google_iam_workload_identity_pool.github.workload_identity_pool_id
  workload_identity_pool_provider_id = "github-provider"
  display_name                       = "GitHub Provider"

  attribute_mapping = {
    "google.subject"       = "assertion.sub"
    "attribute.actor"      = "assertion.actor"
    "attribute.repository" = "assertion.repository"
  }

  oidc {
    issuer_uri = "https://token.actions.githubusercontent.com"
  }

  attribute_condition = "assertion.repository == '${var.github_repo}'"
}

# 서비스 계정 바인딩
resource "google_service_account_iam_member" "github_actions" {
  service_account_id = google_service_account.deploy.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.github.name}/attribute.repository/${var.github_repo}"
}
```

### GitHub Actions Workflow

```yaml
jobs:
  deploy:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      id-token: write

    steps:
      - uses: actions/checkout@v4

      - id: auth
        uses: google-github-actions/auth@v2
        with:
          workload_identity_provider: 'projects/123456789/locations/global/workloadIdentityPools/github-pool/providers/github-provider'
          service_account: 'deploy-sa@project-id.iam.gserviceaccount.com'

      - uses: google-github-actions/setup-gcloud@v2

      - run: gcloud storage cp file.txt gs://bucket-name/
```

---

## Predefined Roles Reference

### Storage

| 역할 | 설명 |
|------|------|
| roles/storage.admin | 전체 관리 |
| roles/storage.objectAdmin | 객체 관리 (버킷 제외) |
| roles/storage.objectViewer | 객체 읽기 |
| roles/storage.objectCreator | 객체 생성만 |

### Compute

| 역할 | 설명 |
|------|------|
| roles/compute.admin | 전체 관리 |
| roles/compute.instanceAdmin | 인스턴스 관리 |
| roles/compute.viewer | 읽기 전용 |
| roles/compute.networkAdmin | 네트워크 관리 |

### Cloud Run

| 역할 | 설명 |
|------|------|
| roles/run.admin | 전체 관리 |
| roles/run.developer | 배포 권한 |
| roles/run.invoker | 서비스 호출 |
| roles/run.viewer | 읽기 전용 |

---

## Best Practices

### Least Privilege

- 필요한 최소 권한만 부여
- 프로젝트 수준보다 리소스 수준 바인딩 선호
- 커스텀 역할로 세밀한 권한 제어

### Key Management

- 서비스 계정 키 사용 최소화
- Workload Identity 또는 Workload Identity Federation 사용
- 키 사용 시 정기적 로테이션
- Secret Manager로 키 저장

### Monitoring

- Cloud Audit Logs 활성화
- IAM 변경 알림 설정
- 주기적인 IAM 정책 검토
- Policy Analyzer로 권한 분석
