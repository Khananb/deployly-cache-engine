package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"deployly-cache/api/internal/auth"
	"deployly-cache/api/internal/storage"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CacheHandler struct {
	db      *pgxpool.Pool
	storage *storage.Client
}

func NewCacheHandler(db *pgxpool.Pool, storage *storage.Client) *CacheHandler {
	return &CacheHandler{
		db:      db,
		storage: storage,
	}
}

type PresignRequest struct {
	CacheKey string `json:"cache_key"`
}

type PresignResponse struct {
	URL        string `json:"url"`
	ObjectPath string `json:"object_path"`
}

type CompleteUploadRequest struct {
	CacheKey  string `json:"cache_key"`
	SizeBytes int64  `json:"size_bytes"`
}

// RequestUpload generates a pre-signed URL for the CLI to PUT a cache archive
func (h *CacheHandler) RequestUpload(w http.ResponseWriter, r *http.Request) {
	projectID := auth.GetProjectID(r.Context())
	
	var req PresignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.CacheKey == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Define storage path: project_id/cache_key.tar.zst
	objectPath := projectID.String() + "/" + req.CacheKey + ".tar.zst"

	// Generate 15-minute expiry URL
	presignedURL, err := h.storage.GenerateUploadURL(r.Context(), objectPath, 15*time.Minute)
	if err != nil {
		http.Error(w, "Failed to generate upload URL", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(PresignResponse{
		URL:        presignedURL.String(),
		ObjectPath: objectPath,
	})
}

// RequestDownload generates a pre-signed URL for the CLI to GET a cache archive
func (h *CacheHandler) RequestDownload(w http.ResponseWriter, r *http.Request) {
	projectID := auth.GetProjectID(r.Context())
	cacheKey := r.URL.Query().Get("key")

	if cacheKey == "" {
		http.Error(w, "Missing cache key", http.StatusBadRequest)
		return
	}

	var storagePath string
	query := `SELECT storage_path FROM cache_entries WHERE project_id = $1 AND cache_key = $2`
	err := h.db.QueryRow(r.Context(), query, projectID, cacheKey).Scan(&storagePath)
	if err != nil {
		http.Error(w, "Cache entry not found", http.StatusNotFound)
		return
	}

	// Update hit count
	go h.db.Exec(context.Background(), "UPDATE cache_entries SET hit_count = hit_count + 1 WHERE project_id = $1 AND cache_key = $2", projectID, cacheKey)

	presignedURL, err := h.storage.GenerateDownloadURL(r.Context(), storagePath, 15*time.Minute)
	if err != nil {
		http.Error(w, "Failed to generate download URL", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(PresignResponse{
		URL: presignedURL.String(),
	})
}

// CompleteUpload records metadata in Postgres after the CLI confirms upload success
func (h *CacheHandler) CompleteUpload(w http.ResponseWriter, r *http.Request) {
	projectID := auth.GetProjectID(r.Context())

	var req CompleteUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	objectPath := projectID.String() + "/" + req.CacheKey + ".tar.zst"
	expiresAt := time.Now().AddDate(0, 0, 30) // Default 30-day expiry

	query := `
		INSERT INTO cache_entries (project_id, cache_key, storage_path, size_bytes, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (project_id, cache_key) 
		DO UPDATE SET 
			size_bytes = EXCLUDED.size_bytes,
			expires_at = EXCLUDED.expires_at,
			created_at = CURRENT_TIMESTAMP`

	_, err := h.db.Exec(r.Context(), query, projectID, req.CacheKey, objectPath, req.SizeBytes, expiresAt)
	if err != nil {
		http.Error(w, "Failed to record cache metadata", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
