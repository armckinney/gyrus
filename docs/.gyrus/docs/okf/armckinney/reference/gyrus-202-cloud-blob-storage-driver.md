---
id: gyrus-202-cloud-blob-storage-driver
title: 'GYRUS-202: Cloud Blob Storage Driver (S3, Azure, GCS)'
category: technical
type: specification
format: ""
owner_group: armckinney
version: 1
status: completed
last_modified_by: ""
last_updated: 2026-07-23T21:23:12Z
tags:
    - phase-2.1
    - storage-driver
    - s3
    - azure
    - gcs
---

# GYRUS-202: Cloud Blob Storage Driver (S3, Azure, GCS)

## 1. Overview & Objective
Implement a cloud object storage `DocumentStore` provider under `internal/provider/blob` using Go Cloud Development Kit (`gocloud.dev/blob`). Enables cloud-native storage of OKF Markdown document bundles across AWS S3 (`s3://`), Azure Blob (`azblob://`), and Google Cloud Storage (`gs://`).

## 2. Requirements & Constraints
- Must implement `gyrus.DocumentStore` interface.
- Must delegate credential resolution automatically to cloud SDK default chains (`AWS_ACCESS_KEY_ID`/`AWS_PROFILE`, `AZURE_STORAGE_ACCOUNT`, `GOOGLE_APPLICATION_CREDENTIALS` / CLI logins) with zero required hardcoded credentials.
- Key format: `<prefix>/<owner_group>/<category>/<id>.md`.
- Must support URL bucket opening (e.g. `blob.OpenBucket(ctx, "s3://my-bucket")`).

## 3. Key Test Verification
- Unit tests using in-memory `fileblob` or mock bucket provider in `internal/provider/blob/blob_test.go`.
- CRUD operations test verifying store/get/update/delete operations.
