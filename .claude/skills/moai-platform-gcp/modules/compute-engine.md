# Compute Engine Module

## Overview

Google Compute Engine은 확장 가능한 가상 머신 서비스로 ETL 워커, 배치 처리, 장기 실행 프로세스에 적합합니다.

---

## Machine Types

### General Purpose (E2)

비용 효율적인 범용 워크로드에 적합합니다.

| 타입 | vCPU | 메모리 | 사용 사례 |
|------|------|--------|----------|
| e2-micro | 0.25 | 1GB | 테스트, 개발 |
| e2-small | 0.5 | 2GB | 경량 서비스 |
| e2-medium | 1 | 4GB | 소규모 애플리케이션 |
| e2-standard-2 | 2 | 8GB | 일반 워크로드 |
| e2-standard-4 | 4 | 16GB | 중규모 ETL |
| e2-standard-8 | 8 | 32GB | 대규모 처리 |

### Compute Optimized (C2)

CPU 집약적 워크로드에 적합합니다.

| 타입 | vCPU | 메모리 | 사용 사례 |
|------|------|--------|----------|
| c2-standard-4 | 4 | 16GB | 고성능 컴퓨팅 |
| c2-standard-8 | 8 | 32GB | 병렬 처리 |
| c2-standard-16 | 16 | 64GB | 대규모 배치 |

### Memory Optimized (M2)

메모리 집약적 워크로드에 적합합니다.

| 타입 | vCPU | 메모리 | 사용 사례 |
|------|------|--------|----------|
| m2-ultramem-208 | 208 | 5.75TB | 인메모리 DB |
| m2-megamem-416 | 416 | 5.75TB | SAP HANA |

### Custom Machine Types

정확한 리소스 요구사항에 맞게 커스텀 타입을 생성합니다.

```hcl
resource "google_compute_instance" "custom" {
  name         = "custom-vm"
  machine_type = "custom-4-8192"  # 4 vCPU, 8GB RAM
  zone         = "asia-northeast3-a"

  # ...
}
```

---

## Preemptible & Spot VMs

최대 80% 비용 절감이 가능한 인터럽트 가능 VM입니다.

### Preemptible VM

```hcl
resource "google_compute_instance" "preemptible" {
  name         = "preemptible-worker"
  machine_type = "e2-standard-4"
  zone         = "asia-northeast3-a"

  scheduling {
    preemptible       = true
    automatic_restart = false
  }

  # ...
}
```

### Spot VM (권장)

Preemptible의 개선 버전으로 최대 24시간 제한이 없습니다.

```hcl
resource "google_compute_instance" "spot" {
  name         = "spot-worker"
  machine_type = "e2-standard-4"
  zone         = "asia-northeast3-a"

  scheduling {
    provisioning_model = "SPOT"
    preemptible        = true
    automatic_restart  = false
  }

  # ...
}
```

---

## Startup Scripts

인스턴스 시작 시 자동으로 실행되는 스크립트입니다.

### Inline Script

```hcl
resource "google_compute_instance" "with_script" {
  # ...

  metadata_startup_script = <<-EOF
    #!/bin/bash
    set -e

    # 로깅
    exec > >(tee /var/log/startup.log) 2>&1

    # 패키지 설치
    apt-get update
    apt-get install -y docker.io

    # Docker 시작
    systemctl start docker
    systemctl enable docker

    # 애플리케이션 실행
    docker pull gcr.io/project/app:latest
    docker run -d --restart=always gcr.io/project/app:latest
  EOF
}
```

### External Script

```hcl
resource "google_compute_instance" "with_external_script" {
  # ...

  metadata = {
    startup-script-url = "gs://bucket-name/scripts/startup.sh"
  }
}
```

---

## Instance Templates

재사용 가능한 인스턴스 구성입니다.

```hcl
resource "google_compute_instance_template" "worker" {
  name_prefix  = "worker-"
  machine_type = "e2-standard-4"
  region       = "asia-northeast3"

  disk {
    source_image = "ubuntu-os-cloud/ubuntu-2204-lts"
    disk_size_gb = 50
    disk_type    = "pd-ssd"
    auto_delete  = true
    boot         = true
  }

  network_interface {
    network = "default"
    access_config {}
  }

  service_account {
    email  = google_service_account.worker.email
    scopes = ["cloud-platform"]
  }

  metadata_startup_script = file("${path.module}/startup.sh")

  labels = {
    role = "worker"
  }

  lifecycle {
    create_before_destroy = true
  }
}
```

---

## Managed Instance Groups

자동 복구 및 확장이 가능한 인스턴스 그룹입니다.

```hcl
resource "google_compute_instance_group_manager" "workers" {
  name               = "worker-group"
  base_instance_name = "worker"
  zone               = "asia-northeast3-a"
  target_size        = 3

  version {
    instance_template = google_compute_instance_template.worker.id
  }

  named_port {
    name = "http"
    port = 8080
  }

  auto_healing_policies {
    health_check      = google_compute_health_check.worker.id
    initial_delay_sec = 300
  }

  update_policy {
    type                  = "PROACTIVE"
    minimal_action        = "REPLACE"
    max_surge_fixed       = 3
    max_unavailable_fixed = 0
  }
}
```

### Autoscaler

```hcl
resource "google_compute_autoscaler" "workers" {
  name   = "worker-autoscaler"
  zone   = "asia-northeast3-a"
  target = google_compute_instance_group_manager.workers.id

  autoscaling_policy {
    min_replicas    = 1
    max_replicas    = 10
    cooldown_period = 60

    cpu_utilization {
      target = 0.7
    }
  }
}
```

---

## Network Configuration

### Firewall Rules

```hcl
resource "google_compute_firewall" "allow_internal" {
  name    = "allow-internal"
  network = "default"

  allow {
    protocol = "tcp"
    ports    = ["0-65535"]
  }

  allow {
    protocol = "udp"
    ports    = ["0-65535"]
  }

  allow {
    protocol = "icmp"
  }

  source_ranges = ["10.128.0.0/9"]
}

resource "google_compute_firewall" "allow_ssh" {
  name    = "allow-ssh"
  network = "default"

  allow {
    protocol = "tcp"
    ports    = ["22"]
  }

  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["ssh-enabled"]
}
```

### Private IP Only

```hcl
resource "google_compute_instance" "private" {
  # ...

  network_interface {
    network = "default"
    # access_config 생략으로 외부 IP 없음
  }
}
```

---

## Health Checks

```hcl
resource "google_compute_health_check" "http" {
  name               = "http-health-check"
  check_interval_sec = 30
  timeout_sec        = 10
  healthy_threshold  = 2
  unhealthy_threshold = 3

  http_health_check {
    port         = 8080
    request_path = "/health"
  }
}

resource "google_compute_health_check" "tcp" {
  name               = "tcp-health-check"
  check_interval_sec = 30
  timeout_sec        = 10

  tcp_health_check {
    port = 8080
  }
}
```

---

## Best Practices

### Cost Optimization

- Spot VM으로 최대 80% 절감
- 커스텀 머신 타입으로 정확한 리소스 할당
- Committed Use Discounts 활용
- 사용하지 않는 인스턴스 자동 중지

### Security

- 서비스 계정으로 인증
- VPC 네트워크 격리
- OS Login으로 SSH 접근 관리
- Shielded VM으로 부팅 무결성 보장

### Reliability

- Managed Instance Groups로 자동 복구
- 멀티 존 배포로 고가용성
- 스냅샷으로 백업
- 모니터링 및 알림 설정
