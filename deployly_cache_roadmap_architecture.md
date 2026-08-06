# Deployly Cache Engine: Technical Specification & Roadmap

## 1. System Architecture
Deployly Cache Engine is designed as a lightweight, high-performance dependency caching service for CI/CD pipelines.

### Components
- **CLI (Go):** Client-side binary for `save` and `restore` operations.
- **API Server (Go):** Orchestrates authentication, metadata management, and S3 pre-signed URL generation.
- **Database (PostgreSQL):** Stores user accounts, API keys, cache keys, and metadata (size, hits, expiry).
- **Storage (MinIO/S3):** Content-addressable storage for zstd-compressed cache archives.

## 2. Core Entities
- **Project:** Namespace for caches (e.g., `github.com/org/repo`).
- **Cache Entry:** A specific version of a cache, identified by a unique key (hash of lockfiles).
- **API Key:** Scoped access tokens for CI runners.

## 3. Initial Project Structure
```text
deployly-cache/
├── api/              # Go REST API
│   ├── internal/     # Private library code
│   │   ├── auth/     # Middleware & API Key validation
│   │   ├── db/       # Postgres migrations & queries
│   │   └── storage/  # MinIO/S3 client logic
│   └── main.go       # API Entry point
├── cli/              # Go CLI Tool
│   ├── cmd/          # CLI commands (save/restore)
│   └── pkg/          # Reusable CLI logic (compression/hashing)
├── scripts/          # Deployment & setup scripts
└── docker-compose.yml # Local dev environment (Postgres + MinIO)
```

## 4. Development Rules (Confirmed)
1. **Zero Placeholders:** Every code block will be feature-complete.
2. **SOLID & Clean Architecture:** Separation of concerns between transport, business logic, and data layers.
3. **VPS Optimized:** Focus on low memory footprint and efficient I/O.
4. **Copy-Paste Ready:** Explicit file paths and schemas provided.
