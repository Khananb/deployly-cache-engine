package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type contextKey string

const ProjectIDKey contextKey = "project_id"

// Middleware provides API Key validation against the database
func Middleware(pool *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Unauthorized: Missing or invalid API key", http.StatusUnauthorized)
				return
			}

			apiKey := strings.TrimPrefix(authHeader, "Bearer ")
			
			// Hash the key to match database storage (SHA-256 for this implementation)
			hash := sha256.Sum256([]byte(apiKey))
			keyHash := hex.EncodeToString(hash[:])

			var projectID uuid.UUID
			query := `
				UPDATE api_keys 
				SET last_used_at = CURRENT_TIMESTAMP 
				WHERE key_hash = $1 
				RETURNING project_id`

			err := pool.QueryRow(r.Context(), query, keyHash).Scan(&projectID)
			if err != nil {
				// We don't distinguish between DB errors and invalid keys for security
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Inject project_id into context for downstream handlers
			ctx := context.WithValue(r.Context(), ProjectIDKey, projectID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetProjectID retrieves the project UUID from the request context
func GetProjectID(ctx context.Context) uuid.UUID {
	id, _ := ctx.Value(ProjectIDKey).(uuid.UUID)
	return id
}
