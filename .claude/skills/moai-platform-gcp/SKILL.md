---
name: "moai-platform-gcp"
description: "Google Cloud Platform specialist covering Cloud Storage (GCS), Compute Engine, Cloud Run, and IAM. Use when building ETL pipelines on GCP, deploying containers to Cloud Run, managing storage buckets, or configuring service accounts."
version: 1.0.0
category: "platform"
modularized: true
user-invocable: false
tags: ['gcp', 'google-cloud', 'cloud-storage', 'gcs', 'compute-engine', 'cloud-run', 'iam', 'terraform', 'go']
context7-libraries: "/google-cloud-go/google-cloud-go, /hashicorp/terraform-provider-google"
related-skills: "moai-platform-vercel, moai-platform-railway, moai-lang-go"
updated: 2026-01-18
status: "active"
allowed-tools:
  - Read
  - Write
  - Bash
  - Grep
  - Glob
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
triggers:
  keywords:
    en: ["gcp", "google cloud", "cloud storage", "gcs", "compute engine", "cloud run", "gcloud", "gsutil"]
    ko: ["GCP", "구글클라우드", "클라우드스토리지", "컴퓨트엔진", "클라우드런"]
---

# moai-platform-gcp: Google Cloud Platform Specialist

## Quick Reference

GCP Infrastructure Focus: Comprehensive cloud platform covering object storage with Cloud Storage, virtual machines with Compute Engine, serverless containers with Cloud Run, and identity management with IAM and service accounts.

### Core Capabilities

Cloud Storage (GCS) provides highly durable object storage with global edge caching for CDN-like performance. Features include multi-regional and regional storage classes, lifecycle policies for automatic data management, uniform bucket-level access for simplified permissions, and streaming uploads for large file handling.

Compute Engine delivers scalable virtual machines with custom machine types and preemptible instances for cost optimization. Features include startup scripts for automated provisioning, instance templates for reproducible deployments, and managed instance groups for auto-scaling.

Cloud Run enables serverless container deployments with automatic scaling from zero to thousands of instances. Features include custom domains with managed SSL, traffic splitting for gradual rollouts, and VPC connectivity for private networking.

IAM and Service Accounts provide fine-grained access control with service accounts for application authentication. Features include Workload Identity for GKE integration, custom roles for least-privilege access, and organization-level policies for governance.

### Quick Decision Guide

Choose GCP Cloud Storage when large-scale object storage is required, data needs to be accessed globally with low latency, lifecycle management is important for cost optimization, or integration with other GCP services is needed.

Choose Compute Engine when full control over VM configuration is required, custom machine types are needed for specific workloads, preemptible instances can reduce costs significantly, or complex networking configurations are required.

Choose Cloud Run when containerized applications need serverless deployment, auto-scaling from zero is cost-effective for variable workloads, quick deployments without infrastructure management are preferred, or HTTP-based services with simple scaling requirements are needed.

---

## Module Index

This skill uses modular documentation for progressive disclosure. Load modules as needed:

### Cloud Storage Module

Module at modules/cloud-storage.md covers bucket creation and configuration, object upload with streaming and resumable uploads, download and listing operations, lifecycle policies and storage classes, uniform bucket-level access and IAM, signed URLs for temporary access, and transfer service for large migrations.

### Compute Engine Module

Module at modules/compute-engine.md covers VM instance creation and management, machine type selection and custom types, boot disk configuration and snapshots, startup scripts and metadata, network configuration and firewall rules, instance templates and managed instance groups, and preemptible and spot instances.

### Cloud Run Module

Module at modules/cloud-run.md covers container deployment from Artifact Registry, auto-scaling configuration and concurrency, custom domain mapping with managed SSL, traffic splitting and revisions, VPC connectors for private networking, and Cloud Run Jobs for batch processing.

### IAM Module

Module at modules/iam.md covers service account creation and key management, role binding and custom roles, Workload Identity for Kubernetes, organization policies and constraints, and cross-project resource access.

---

## Implementation Guide

### Go SDK Quick Start

For Go applications using Cloud Storage, import the storage package from cloud.google.com/go/storage. Create a new storage client with context using storage.NewClient. Handle the error appropriately. For bucket operations, use client.Bucket to get a bucket handle, then NewWriter for uploads and NewReader for downloads. Set ContentType on the writer before writing data, and always close the writer to complete the upload.

### Terraform Quick Start

For Terraform infrastructure as code, configure the google provider with your project and region. Define google_storage_bucket resources with name, location, and uniform_bucket_level_access enabled. Add lifecycle_rule blocks with condition for age and action for deletion or storage class transitions. For Compute Engine, define google_compute_instance resources with machine_type, zone, boot_disk with initialize_params for the image, and network_interface with access_config for external IP.

### gcloud CLI Essentials

For Cloud Storage operations, use gsutil mb to create buckets, gsutil cp for copying files, gsutil ls for listing, and gsutil rm for deletion. For Compute Engine, use gcloud compute instances create with machine-type, image-family, and zone flags. For Cloud Run, use gcloud run deploy with source, region, and allow-unauthenticated flags.

### Authentication Patterns

For local development, use gcloud auth application-default login to set up credentials. For production environments, create a service account with gcloud iam service-accounts create, grant roles with gcloud projects add-iam-policy-binding, and create keys with gcloud iam service-accounts keys create. For containerized applications, use Workload Identity to avoid key management.

---

## Advanced Patterns

### ETL Pipeline Architecture

For ETL pipelines on GCP, use Cloud Storage as the data lake for raw and processed data. Configure lifecycle policies to transition data to cheaper storage classes over time. Use Compute Engine or Cloud Run for processing depending on workload characteristics. Implement service accounts with minimal required permissions following least-privilege principles.

### Multi-Region Deployment

For high availability deployments, use multi-regional Cloud Storage buckets for data durability. Deploy Cloud Run services in multiple regions with Cloud Load Balancing. Use global IAM policies for consistent access control across regions.

### Context7 Integration

Use Context7 MCP tools for latest documentation:

Step 1: Resolve library ID using mcp__context7__resolve-library-id with "google-cloud-go" for Go SDK or "terraform-provider-google" for Terraform.

Step 2: Fetch documentation using mcp__context7__get-library-docs with resolved ID, specific topic like "cloud storage bucket" or "compute instance", and appropriate token allocation of 5000 to 10000 tokens.

---

## Best Practices

### Security

Always enable uniform bucket-level access for Cloud Storage buckets to simplify permission management. Use service accounts instead of user accounts for application authentication. Implement least-privilege access with custom roles when predefined roles grant too much access. Enable Cloud Audit Logs for compliance and security monitoring.

### Cost Optimization

Use lifecycle policies to automatically transition objects to cheaper storage classes like Nearline or Coldline. Consider preemptible or spot VMs for fault-tolerant workloads to reduce costs by up to 80%. Configure Cloud Run minimum instances to zero for infrequently used services. Use committed use discounts for predictable workloads.

### Performance

Choose regional storage for data that is accessed primarily from one region. Use resumable uploads for files larger than 5MB to handle network interruptions. Configure Cloud Run concurrency appropriately based on application characteristics. Use Cloud CDN with Cloud Storage for frequently accessed public content.

---

## Works Well With

- moai-lang-go for Go patterns with Cloud Storage SDK
- moai-platform-railway for alternative container deployment
- moai-platform-vercel for frontend deployment with GCP backend
- moai-foundation-quality for infrastructure validation and testing
- moai-domain-backend for API development on GCP

---

## Additional Resources

- Reference Guide at reference.md provides complete gcloud CLI commands and Terraform configuration options
- Code Examples at examples.md provides production-ready Go SDK and Terraform code snippets

---

Status: Production Ready
Version: 1.0.0
Updated: 2026-01-18
