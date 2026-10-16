---
title: 'GYRUS-202: Cloud Blob Storage Driver (S3, Azure, GCS)'
id: spec-202-cloud-blob-storage-driver
type: specification
format: markdown
category: technical
version: 2
status: completed
last_updated: 2026-10-16T06:40:18Z
last_modified_by: ""
tenant: root
tags:
    - phase-2.1
    - storage-driver
    - s3
    - azure
    - gcs
---

# Cloud blob storage driver

The blob ConceptStore uses Go Cloud object buckets for S3, Azure Blob, GCS, and fileblob. The storage factory retains existing credential discovery and provider aliases. Each cloud profile now defaults to direct-scan OKF index/graph/search.

## Canonical layout and scanner

Keys are <prefix>/concepts/<tenant>/reference/<id>.md or <prefix>/concepts/<tenant>/projects/<project>/<id>.md. Prefix boundaries are normalized. Category-based legacy keys are not supported. Metadata scope/project derive from keys; all concept CRUD paths use the same serializer and typed-link schema.

ConceptScanner enumerates only this concept collection, validates payloads, excludes schemas/unrelated documentation, and reports invalid records and duplicate IDs. Callbacks see detached concepts. OKF Sync scans this provider rather than a local filesystem path.

## Concurrency

Read content and ETag/generation together. Create requires nonexistence; update/delete require the observed token plus any positive expected-version check. S3 and Azure use ETags; GCS uses object generations. Fileblob uses a root-wide interprocess lock and complete-payload atomic replacement. Opaque injected buckets without safe mutation capabilities remain readable but reject mutation.

Precondition failures report typed concurrency conflicts without overwriting the winner. Unsupported conditional mutation hooks fail explicitly. There is no cross-object transaction, collection-wide atomic snapshot, or distributed cross-key uniqueness reservation. Concurrent same-ID creates at different keys are detected as duplicate IDs by subsequent scans.

## Verification

Use fileblob integration tests and credential-free local HTTP servers for cloud reader tokens, mutation conditions, stale update/delete rejection, and unsupported-driver behavior. Live cloud accounts are optional for routine validation.
