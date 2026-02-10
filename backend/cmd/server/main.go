package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

	"github.com/sherlockos/backend/internal/api"
	"github.com/sherlockos/backend/internal/clients"
	"github.com/sherlockos/backend/internal/db"
	"github.com/sherlockos/backend/internal/queue"
	"github.com/sherlockos/backend/internal/workers"
	"github.com/sherlockos/backend/pkg/config"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize database connection
	database, err := db.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Initialize job queue (Redis with fallback to in-memory)
	jobQueue, err := queue.NewWithFallback(cfg.RedisURL)
	if err != nil {
		log.Printf("Warning: Failed to initialize queue: %v (jobs will not be processed)", err)
	} else {
		defer jobQueue.Close()
		if cfg.RedisURL != "" {
			log.Println("Redis queue initialized")
		} else {
			log.Println("In-memory queue initialized (Redis URL not configured)")
		}
	}

	// Initialize storage client (needed by AI clients for image fetching)
	var storageClient clients.StorageClient
	if cfg.SupabaseURL != "" && cfg.SupabaseSecretKey != "" {
		storageClient = clients.NewSupabaseStorageClient(cfg.SupabaseURL, cfg.SupabaseSecretKey)
		log.Println("Supabase storage client initialized")
	}

	// Initialize AI clients and workers if Gemini API key is available
	var workerManager *workers.Manager
	if cfg.GeminiAPIKey != "" {
		// Initialize Gemini API clients
		reasoningClient := clients.NewGeminiReasoningClient(cfg.GeminiAPIKey)
		profileClient := clients.NewGeminiProfileClient(cfg.GeminiAPIKey)
		imageGenClient := clients.NewGeminiImageGenClient(cfg.GeminiAPIKey, storageClient)

		// Initialize worker manager
		workerManager = workers.NewManager(database, jobQueue, workers.DefaultManagerConfig())

		// Register Gemini-based workers (always available when GEMINI_API_KEY is set)
		workerManager.Register(workers.NewReasoningWorker(database, jobQueue, reasoningClient))
		workerManager.Register(workers.NewProfileWorker(database, jobQueue, profileClient))
		workerManager.Register(workers.NewImageGenWorker(database, jobQueue, imageGenClient))
		log.Println("Gemini workers registered (reasoning, profile, imagegen)")

		// Initialize reconstruction client (Modal HunyuanWorld-Mirror) - NO MOCK FALLBACK
		if cfg.ModalMirrorURL != "" && storageClient != nil {
			reconstructionClient := clients.NewModalReconstructionClient(cfg.ModalMirrorURL, storageClient)
			workerManager.Register(workers.NewReconstructionWorker(database, jobQueue, reconstructionClient))
			log.Println("Reconstruction worker registered (Modal HunyuanWorld-Mirror)")
		} else {
			log.Println("WARNING: Reconstruction worker DISABLED - MODAL_MIRROR_URL not set or storage not configured")
			log.Println("  → reconstruction jobs will fail with 'service not available' error")
		}

		// Initialize replay client (Modal HY-WorldPlay) - NO MOCK FALLBACK
		if cfg.ModalWorldPlayURL != "" && storageClient != nil {
			replayClient := clients.NewModalReplayClient(cfg.ModalWorldPlayURL, storageClient)
			workerManager.Register(workers.NewReplayWorker(database, jobQueue, replayClient))
			log.Println("Replay worker registered (Modal HY-WorldPlay)")
		} else {
			log.Println("WARNING: Replay worker DISABLED - MODAL_WORLDPLAY_URL not set or storage not configured")
			log.Println("  → replay jobs will fail with 'service not available' error")
		}

		// Register scene analysis worker (Gemini 3 Pro Vision)
		if storageClient != nil {
			sceneAnalysisClient := clients.NewGeminiSceneAnalysisClient(cfg.GeminiAPIKey, storageClient)
			workerManager.Register(workers.NewSceneAnalysisWorker(database, jobQueue, sceneAnalysisClient))
			log.Println("Scene analysis worker registered (Gemini 3 Pro Vision)")
		}

		// Register 3D asset worker (Hunyuan3D-2 via Replicate)
		if cfg.ReplicateAPIToken != "" && storageClient != nil {
			asset3dClient := clients.NewReplicateAsset3DClient(cfg.ReplicateAPIToken, storageClient)
			workerManager.Register(workers.NewAsset3DWorker(database, jobQueue, asset3dClient))
			log.Println("3D asset worker registered (Hunyuan3D-2 via Replicate)")
		} else if cfg.ReplicateAPIToken == "" {
			log.Println("Warning: REPLICATE_API_TOKEN not set, 3D asset worker disabled")
		}

		// Register export worker (HTML report generation)
		workerManager.Register(workers.NewExportWorker(database, jobQueue, storageClient))
		log.Println("Export worker registered (HTML report generation)")

		// Start workers
		workerManager.Start(context.Background())
		log.Println("Workers started")
	} else {
		if cfg.GeminiAPIKey == "" {
			log.Println("Warning: GEMINI_API_KEY not set, AI workers disabled")
		}
	}

	// Initialize router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "Idempotency-Key"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// API routes
	r.Route("/v1", func(r chi.Router) {
		api.RegisterRoutesWithQueue(r, database, jobQueue)

		// Portrait chat route (needs direct access to Gemini client)
		if cfg.GeminiAPIKey != "" {
			imageGenClient := clients.NewGeminiImageGenClient(cfg.GeminiAPIKey, storageClient)
			api.RegisterPortraitRoutes(r, imageGenClient)
		}
	})

	// Create server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Stop workers first
	if workerManager != nil {
		log.Println("Stopping workers...")
		workerManager.Stop()
		log.Println("Workers stopped")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
