package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"deployly-cache/api/internal/auth"
	"deployly-cache/api/internal/db"
	"deployly-cache/api/internal/handlers"
	"deployly-cache/api/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 1. Initialize Database
	database, err := db.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()
	log.Println("Database connection pool established")

	// 2. Initialize Storage (MinIO)
	storageClient, err := storage.NewClient()
	if err != nil {
		log.Fatalf("Failed to initialize storage client: %v", err)
	}
	log.Println("Storage client initialized")

	// 3. Initialize Handlers
	cacheHandler := handlers.NewCacheHandler(database.Pool, storageClient)

	// 4. Initialize Router
	r := chi.NewRouter()

	// Global Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Health Check (Public)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// API Routes (Protected)
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(auth.Middleware(database.Pool))

		r.Get("/check", func(w http.ResponseWriter, r *http.Request) {
			projectID := auth.GetProjectID(r.Context())
			w.Write([]byte("Authenticated for project: " + projectID.String()))
		})
		
		// Cache Operations
		r.Route("/cache", func(r chi.Router) {
			r.Post("/request-upload", cacheHandler.RequestUpload)
			r.Post("/complete-upload", cacheHandler.CompleteUpload)
			r.Get("/restore", cacheHandler.RequestDownload)
		})
	})

	// 5. Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		log.Printf("Deployly Cache API starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen error: %v", err)
		}
	}()

	// Graceful Shutdown
	<-ctx.Done()
	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
