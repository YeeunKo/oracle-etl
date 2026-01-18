# Cloud Run Module

## Overview

Cloud Run은 서버리스 컨테이너 플랫폼으로 HTTP 요청 기반 자동 스케일링을 제공합니다. 프론트엔드 배포, API 서비스, 마이크로서비스에 적합합니다.

---

## Service Configuration

### Basic Service

```hcl
resource "google_cloud_run_v2_service" "api" {
  name     = "api-service"
  location = "asia-northeast3"

  template {
    containers {
      image = "gcr.io/project-id/api:latest"

      ports {
        container_port = 8080
      }

      resources {
        limits = {
          cpu    = "1000m"
          memory = "512Mi"
        }
      }
    }
  }

  traffic {
    type    = "TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST"
    percent = 100
  }
}
```

### Environment Variables

```hcl
template {
  containers {
    image = "gcr.io/project-id/api:latest"

    # 일반 환경 변수
    env {
      name  = "ENV"
      value = "production"
    }

    env {
      name  = "LOG_LEVEL"
      value = "info"
    }

    # Secret Manager에서 값 가져오기
    env {
      name = "DATABASE_URL"
      value_source {
        secret_key_ref {
          secret  = "database-url"
          version = "latest"
        }
      }
    }

    env {
      name = "API_KEY"
      value_source {
        secret_key_ref {
          secret  = "api-key"
          version = "1"  # 특정 버전
        }
      }
    }
  }
}
```

---

## Scaling Configuration

### Auto Scaling

```hcl
template {
  containers {
    image = "gcr.io/project-id/api:latest"

    resources {
      limits = {
        cpu    = "2000m"
        memory = "1Gi"
      }
      cpu_idle = true  # CPU 유휴 시 스케일 다운
    }
  }

  scaling {
    min_instance_count = 0   # 0으로 스케일 다운
    max_instance_count = 100 # 최대 100 인스턴스
  }

  # 요청당 동시성
  max_instance_request_concurrency = 80
}
```

### Always-On (최소 인스턴스)

콜드 스타트 방지를 위해 최소 인스턴스를 유지합니다.

```hcl
scaling {
  min_instance_count = 1   # 항상 1개 이상 유지
  max_instance_count = 10
}
```

---

## Health Checks

### Startup Probe

컨테이너가 시작될 때까지 기다립니다.

```hcl
template {
  containers {
    image = "gcr.io/project-id/api:latest"

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
  }
}
```

### Liveness Probe

실행 중인 컨테이너의 상태를 확인합니다.

```hcl
template {
  containers {
    image = "gcr.io/project-id/api:latest"

    liveness_probe {
      http_get {
        path = "/health"
        port = 8080
      }
      initial_delay_seconds = 10
      timeout_seconds       = 5
      period_seconds        = 30
      failure_threshold     = 3
    }
  }
}
```

---

## Traffic Management

### Blue-Green Deployment

```hcl
resource "google_cloud_run_v2_service" "api" {
  name     = "api-service"
  location = "asia-northeast3"

  template {
    revision = "api-service-v2"  # 새 리비전
    containers {
      image = "gcr.io/project-id/api:v2"
    }
  }

  # 점진적 롤아웃
  traffic {
    type    = "TRAFFIC_TARGET_ALLOCATION_TYPE_REVISION"
    revision = "api-service-v1"
    percent  = 90
  }

  traffic {
    type    = "TRAFFIC_TARGET_ALLOCATION_TYPE_REVISION"
    revision = "api-service-v2"
    percent  = 10
  }
}
```

### gcloud 트래픽 분할

```bash
# 새 리비전으로 10% 트래픽
gcloud run services update-traffic api-service \
    --to-revisions=api-service-v2=10 \
    --region=asia-northeast3

# 점진적 증가
gcloud run services update-traffic api-service \
    --to-revisions=api-service-v2=50 \
    --region=asia-northeast3

# 100% 전환
gcloud run services update-traffic api-service \
    --to-latest \
    --region=asia-northeast3
```

---

## VPC Connectivity

### VPC Connector

프라이빗 리소스에 접근합니다.

```hcl
resource "google_vpc_access_connector" "connector" {
  name          = "cloudrun-connector"
  region        = "asia-northeast3"
  network       = "default"
  ip_cidr_range = "10.8.0.0/28"
  machine_type  = "e2-micro"

  min_instances = 2
  max_instances = 10
}

resource "google_cloud_run_v2_service" "api" {
  name     = "api-service"
  location = "asia-northeast3"

  template {
    containers {
      image = "gcr.io/project-id/api:latest"
    }

    vpc_access {
      connector = google_vpc_access_connector.connector.id
      egress    = "PRIVATE_RANGES_ONLY"  # 또는 "ALL_TRAFFIC"
    }
  }
}
```

---

## Custom Domains

### Domain Mapping

```hcl
resource "google_cloud_run_domain_mapping" "api" {
  name     = "api.example.com"
  location = "asia-northeast3"

  metadata {
    namespace = "project-id"
  }

  spec {
    route_name = google_cloud_run_v2_service.api.name
  }
}
```

### gcloud 도메인 매핑

```bash
# 도메인 매핑 생성
gcloud run domain-mappings create \
    --service=api-service \
    --domain=api.example.com \
    --region=asia-northeast3

# 도메인 매핑 확인
gcloud run domain-mappings describe \
    --domain=api.example.com \
    --region=asia-northeast3
```

---

## Cloud Run Jobs

배치 작업을 위한 일회성 컨테이너 실행입니다.

```hcl
resource "google_cloud_run_v2_job" "etl_job" {
  name     = "etl-batch-job"
  location = "asia-northeast3"

  template {
    template {
      containers {
        image = "gcr.io/project-id/etl-worker:latest"

        env {
          name  = "BATCH_SIZE"
          value = "1000"
        }

        resources {
          limits = {
            cpu    = "2000m"
            memory = "2Gi"
          }
        }
      }

      timeout = "3600s"  # 1시간 타임아웃
      max_retries = 3
    }

    parallelism = 10  # 병렬 작업 수
    task_count  = 100 # 총 작업 수
  }
}
```

### Job 실행

```bash
# Job 실행
gcloud run jobs execute etl-batch-job --region=asia-northeast3

# 실행 상태 확인
gcloud run jobs executions list --job=etl-batch-job --region=asia-northeast3

# 특정 실행 로그 확인
gcloud run jobs executions logs read EXECUTION_NAME --region=asia-northeast3
```

---

## Authentication

### Public Access

```hcl
resource "google_cloud_run_service_iam_member" "public" {
  location = google_cloud_run_v2_service.api.location
  service  = google_cloud_run_v2_service.api.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}
```

### Authenticated Access

```hcl
# 특정 서비스 계정만 허용
resource "google_cloud_run_service_iam_member" "invoker" {
  location = google_cloud_run_v2_service.api.location
  service  = google_cloud_run_v2_service.api.name
  role     = "roles/run.invoker"
  member   = "serviceAccount:${google_service_account.caller.email}"
}
```

### 인증된 요청

```go
import (
    "context"
    "net/http"

    "google.golang.org/api/idtoken"
)

func callCloudRun(ctx context.Context, url string) (*http.Response, error) {
    client, err := idtoken.NewClient(ctx, url)
    if err != nil {
        return nil, err
    }

    return client.Get(url)
}
```

---

## Best Practices

### Performance

- 적절한 동시성 설정 (기본 80)
- 콜드 스타트 최소화 (최소 인스턴스 설정)
- 컨테이너 이미지 최적화 (경량 베이스 이미지)
- CPU 할당 최적화 (요청 처리 중에만)

### Security

- 최소 권한 서비스 계정 사용
- Secret Manager로 비밀 관리
- VPC Connector로 네트워크 격리
- IAM으로 접근 제어

### Cost Optimization

- min_instances=0으로 유휴 비용 절감
- cpu_idle=true로 CPU 비용 최적화
- 적절한 리소스 한도 설정
- 동시성 최적화로 인스턴스 수 최소화
