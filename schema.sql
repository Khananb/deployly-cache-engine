# Deployly Cache Engine: Database Schema

-- Enable UUID extension for secure, non-sequential identifiers
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Projects: Namespaces for caches (e.g., github.com/org/repo)
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- API Keys: Scoped access tokens for CI runners
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    key_hash TEXT NOT NULL UNIQUE, -- Scrypt or Argon2 hash of the actual key
    name TEXT NOT NULL,            -- Descriptive name (e.g., "GitHub Action - Main")
    scopes TEXT[] DEFAULT '{read,write}',
    last_used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Cache Entries: Metadata for zstd-compressed archives stored in MinIO/S3
CREATE TABLE cache_entries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    cache_key TEXT NOT NULL,       -- Unique hash of lockfiles (e.g., "go-mod-{{hash}}")
    storage_path TEXT NOT NULL,    -- Path in S3 bucket (e.g., "project-id/cache-key.tar.zst")
    size_bytes BIGINT NOT NULL,
    hit_count INTEGER DEFAULT 0,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Ensure cache keys are unique within a project
    UNIQUE(project_id, cache_key)
);

-- Indexes for performance
CREATE INDEX idx_cache_entries_project_key ON cache_entries(project_id, cache_key);
CREATE INDEX idx_api_keys_hash ON api_keys(key_hash);
