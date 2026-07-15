---
name: Docker Build
description: Guidelines and commands for building Docker images.
applyTo:
  - "containers/**"
  - "Dockerfile"
  - "Dockerfile.dev"
---

# Docker Build Instructions

- Always use the current directory (`.`) as build context.
- Set container build timeouts to 60+ minutes.
- Verify Docker daemon connectivity: `docker ps`.
