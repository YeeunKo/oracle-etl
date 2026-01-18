# Cloud Storage (GCS) Module

## Overview

Google Cloud Storage는 고가용성, 고내구성 객체 스토리지 서비스로 ETL 파이프라인의 데이터 레이크로 적합합니다.

---

## Bucket Configuration

### Storage Classes

| 클래스 | 최소 저장 기간 | 검색 비용 | 사용 사례 |
|--------|---------------|----------|----------|
| STANDARD | 없음 | 없음 | 자주 액세스되는 핫 데이터 |
| NEARLINE | 30일 | GB당 $0.01 | 월 1회 미만 액세스 |
| COLDLINE | 90일 | GB당 $0.02 | 분기 1회 미만 액세스 |
| ARCHIVE | 365일 | GB당 $0.05 | 백업, 아카이브 |

### Location Types

- **Regional**: 단일 리전, 최저 비용, 리전 내 액세스에 최적
- **Dual-Region**: 두 리전 간 자동 복제, 고가용성
- **Multi-Regional**: 대륙 내 여러 리전에 분산, 전역 액세스에 최적

### Uniform Bucket-Level Access

균일 버킷 수준 액세스를 활성화하면 ACL이 비활성화되고 IAM만으로 권한을 관리합니다.

```hcl
resource "google_storage_bucket" "example" {
  name     = "my-bucket"
  location = "ASIA-NORTHEAST3"

  uniform_bucket_level_access = true
}
```

---

## Lifecycle Policies

### Age-Based Deletion

```hcl
lifecycle_rule {
  condition {
    age = 30  # 30일 후 삭제
  }
  action {
    type = "Delete"
  }
}
```

### Storage Class Transition

```hcl
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
```

### Version-Based Cleanup

```hcl
# 버전 관리 활성화
versioning {
  enabled = true
}

# 이전 버전 3개만 유지
lifecycle_rule {
  condition {
    num_newer_versions = 3
    with_state         = "ARCHIVED"
  }
  action {
    type = "Delete"
  }
}
```

### Prefix-Based Rules

```hcl
lifecycle_rule {
  condition {
    age                   = 7
    matches_prefix        = ["temp/", "logs/"]
  }
  action {
    type = "Delete"
  }
}
```

---

## Go SDK Patterns

### Resumable Upload

대용량 파일 업로드 시 네트워크 중단에 대응합니다.

```go
func resumableUpload(ctx context.Context, client *storage.Client, bucket, object string, r io.Reader) error {
    wc := client.Bucket(bucket).Object(object).NewWriter(ctx)

    // 청크 크기 설정 (기본 16MB)
    wc.ChunkSize = 8 * 1024 * 1024 // 8MB

    // 재시도 가능한 청크 쓰기
    if _, err := io.Copy(wc, r); err != nil {
        wc.Close()
        return fmt.Errorf("io.Copy: %w", err)
    }

    if err := wc.Close(); err != nil {
        return fmt.Errorf("Writer.Close: %w", err)
    }

    return nil
}
```

### Parallel Download

대용량 파일을 병렬로 다운로드합니다.

```go
func parallelDownload(ctx context.Context, client *storage.Client, bucket, object string, numWorkers int) ([]byte, error) {
    obj := client.Bucket(bucket).Object(object)

    // 객체 속성 조회
    attrs, err := obj.Attrs(ctx)
    if err != nil {
        return nil, err
    }

    size := attrs.Size
    chunkSize := size / int64(numWorkers)

    result := make([]byte, size)
    var wg sync.WaitGroup
    errCh := make(chan error, numWorkers)

    for i := 0; i < numWorkers; i++ {
        start := int64(i) * chunkSize
        end := start + chunkSize
        if i == numWorkers-1 {
            end = size
        }

        wg.Add(1)
        go func(start, end int64) {
            defer wg.Done()

            rc, err := obj.NewRangeReader(ctx, start, end-start)
            if err != nil {
                errCh <- err
                return
            }
            defer rc.Close()

            if _, err := io.ReadFull(rc, result[start:end]); err != nil {
                errCh <- err
            }
        }(start, end)
    }

    wg.Wait()
    close(errCh)

    if err := <-errCh; err != nil {
        return nil, err
    }

    return result, nil
}
```

### Signed URLs

임시 액세스를 위한 서명된 URL을 생성합니다.

```go
func generateV4SignedURL(bucket, object string, expiration time.Duration) (string, error) {
    ctx := context.Background()
    client, err := storage.NewClient(ctx)
    if err != nil {
        return "", err
    }
    defer client.Close()

    opts := &storage.SignedURLOptions{
        Scheme:  storage.SigningSchemeV4,
        Method:  "GET",
        Expires: time.Now().Add(expiration),
    }

    url, err := client.Bucket(bucket).SignedURL(object, opts)
    if err != nil {
        return "", err
    }

    return url, nil
}
```

### Batch Operations

여러 객체를 효율적으로 처리합니다.

```go
func batchDelete(ctx context.Context, client *storage.Client, bucket string, objects []string) error {
    bkt := client.Bucket(bucket)

    var wg sync.WaitGroup
    errCh := make(chan error, len(objects))

    // 동시성 제한
    sem := make(chan struct{}, 10)

    for _, obj := range objects {
        wg.Add(1)
        go func(objectName string) {
            defer wg.Done()
            sem <- struct{}{}
            defer func() { <-sem }()

            if err := bkt.Object(objectName).Delete(ctx); err != nil {
                errCh <- fmt.Errorf("delete %s: %w", objectName, err)
            }
        }(obj)
    }

    wg.Wait()
    close(errCh)

    var errs []error
    for err := range errCh {
        errs = append(errs, err)
    }

    if len(errs) > 0 {
        return fmt.Errorf("batch delete errors: %v", errs)
    }

    return nil
}
```

---

## CORS Configuration

웹 애플리케이션에서 직접 업로드를 허용합니다.

```json
[
  {
    "origin": ["https://example.com"],
    "method": ["GET", "PUT", "POST", "DELETE"],
    "responseHeader": ["Content-Type", "Content-Length"],
    "maxAgeSeconds": 3600
  }
]
```

```bash
gsutil cors set cors.json gs://bucket-name
```

---

## Notification (Pub/Sub)

객체 변경 시 Pub/Sub 알림을 설정합니다.

```hcl
resource "google_storage_notification" "notification" {
  bucket         = google_storage_bucket.bucket.name
  payload_format = "JSON_API_V1"
  topic          = google_pubsub_topic.topic.id

  event_types = [
    "OBJECT_FINALIZE",
    "OBJECT_METADATA_UPDATE"
  ]

  custom_attributes = {
    environment = "production"
  }
}
```

---

## Best Practices

### Naming Conventions

- 전역적으로 고유한 이름 사용
- 프로젝트 ID를 접두사로 사용 권장
- 소문자, 숫자, 하이픈만 사용

### Security

- 균일 버킷 수준 액세스 활성화
- 공개 액세스 방지 설정
- 데이터 암호화 (기본 활성화, CMEK 선택)

### Performance

- 대용량 파일은 청크 업로드 사용
- 병렬 전송으로 처리량 향상
- 리전 선택 시 데이터 소비 위치 고려

### Cost Optimization

- 수명 주기 정책으로 자동 스토리지 클래스 전환
- 불필요한 버전 자동 삭제
- 임시 데이터 자동 정리
