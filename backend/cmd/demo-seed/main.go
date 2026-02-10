package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"

	"github.com/sherlockos/backend/internal/clients"
	"github.com/sherlockos/backend/internal/db"
	"github.com/sherlockos/backend/internal/models"
	"github.com/sherlockos/backend/pkg/config"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg := config.Load()

	// Validate required config
	if cfg.SupabaseURL == "" || cfg.SupabaseSecretKey == "" {
		log.Fatal("SUPABASE_URL and SUPABASE_SECRET_KEY are required")
	}
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()

	// Connect to database
	database, err := db.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Create repository
	repo := db.NewRepository(database)

	// Initialize storage client
	storageClient := clients.NewSupabaseStorageClient(cfg.SupabaseURL, cfg.SupabaseSecretKey)

	// Demo images directory
	demoDir := "../demo/images"
	if _, err := os.Stat(demoDir); os.IsNotExist(err) {
		// Try from project root
		demoDir = "../../demo/images"
		if _, err := os.Stat(demoDir); os.IsNotExist(err) {
			log.Fatal("Demo images directory not found. Run from backend directory.")
		}
	}

	log.Println("=== SherlockOS Demo Seed ===")
	log.Println("This will create a demo case with real crime scene images and trigger AI analysis.")
	log.Println("")

	// Step 1: Create a new case
	log.Println("[1/4] Creating demo case...")
	caseID := uuid.New()
	caseModel := &models.Case{
		ID:          caseID,
		Title:       "Demo Crime Scene Investigation",
		Description: "A demonstration case using crime scene photos to test the full SherlockOS pipeline.",
		CreatedAt:   time.Now(),
	}

	err = repo.CreateCase(ctx, caseModel)
	if err != nil {
		log.Fatalf("Failed to create case: %v", err)
	}
	log.Printf("  Created case: %s", caseID)

	// Step 2: Upload demo images to Supabase Storage
	log.Println("[2/4] Uploading demo images to storage...")
	bucket := "assets" // Must match the bucket in scene_analysis_client.go

	// Read and upload each demo image
	files, err := os.ReadDir(demoDir)
	if err != nil {
		log.Fatalf("Failed to read demo directory: %v", err)
	}

	var storageKeys []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Only process image files
		ext := filepath.Ext(file.Name())
		if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
			continue
		}

		filePath := filepath.Join(demoDir, file.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("  Warning: Failed to read %s: %v", file.Name(), err)
			continue
		}

		// Determine content type
		contentType := "image/png"
		if ext == ".jpg" || ext == ".jpeg" {
			contentType = "image/jpeg"
		}

		// Upload to Supabase Storage
		storageKey := fmt.Sprintf("cases/%s/scans/%s", caseID, file.Name())
		err = storageClient.Upload(ctx, bucket, storageKey, data, contentType)
		if err != nil {
			log.Printf("  Warning: Failed to upload %s: %v", file.Name(), err)
			continue
		}

		storageKeys = append(storageKeys, storageKey)
		log.Printf("  Uploaded: %s (%d bytes)", file.Name(), len(data))
	}

	if len(storageKeys) == 0 {
		log.Fatal("No images were uploaded. Check your Supabase Storage configuration.")
	}
	log.Printf("  Total: %d images uploaded", len(storageKeys))

	// Step 3: Create an initial commit to record the uploads
	log.Println("[3/4] Creating upload commit...")
	uploadPayload := map[string]interface{}{
		"storage_keys": storageKeys,
		"count":        len(storageKeys),
		"source":       "demo-seed",
	}
	uploadCommit, err := models.NewCommit(caseID, models.CommitTypeUploadScan, "Demo images uploaded", uploadPayload)
	if err != nil {
		log.Fatalf("Failed to create commit model: %v", err)
	}
	err = repo.CreateCommit(ctx, uploadCommit)
	if err != nil {
		log.Fatalf("Failed to create upload commit: %v", err)
	}
	log.Printf("  Created commit: %s", uploadCommit.ID)

	// Step 4: Create scene_analysis job via API (so it uses the server's queue)
	log.Println("[4/5] Creating scene analysis job via API...")
	sceneJobInput := map[string]interface{}{
		"type": "scene_analysis",
		"input": map[string]interface{}{
			"case_id":    caseID.String(),
			"image_keys": storageKeys,
			"mode":       "full_analysis",
			"query":      "Identify all objects, potential evidence, entry/exit points, and anything suspicious in this crime scene",
		},
	}

	inputJSON, _ := json.Marshal(sceneJobInput)
	apiURL := fmt.Sprintf("http://localhost:%s/v1/cases/%s/jobs", cfg.Port, caseID)
	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(inputJSON))
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Failed to create job via API (is the server running?): %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		log.Fatalf("Failed to create scene_analysis job: %s", string(body))
	}

	var sceneJobResp struct {
		Data struct {
			JobID string `json:"job_id"`
		} `json:"data"`
	}
	json.Unmarshal(body, &sceneJobResp)
	sceneJobID := sceneJobResp.Data.JobID
	log.Printf("  Created job: %s (type: scene_analysis)", sceneJobID)

	// Step 5: Create reconstruction job via API (3D room reconstruction from photos)
	log.Println("[5/5] Creating reconstruction job via API...")
	reconJobInput := map[string]interface{}{
		"type": "reconstruction",
		"input": map[string]interface{}{
			"case_id":         caseID.String(),
			"scan_asset_keys": storageKeys,
		},
	}

	reconInputJSON, _ := json.Marshal(reconJobInput)
	reconReq, err := http.NewRequest("POST", apiURL, bytes.NewReader(reconInputJSON))
	if err != nil {
		log.Fatalf("Failed to create reconstruction request: %v", err)
	}
	reconReq.Header.Set("Content-Type", "application/json")

	reconResp, err := http.DefaultClient.Do(reconReq)
	if err != nil {
		log.Fatalf("Failed to create reconstruction job via API: %v", err)
	}
	defer reconResp.Body.Close()

	reconBody, _ := io.ReadAll(reconResp.Body)
	if reconResp.StatusCode != http.StatusAccepted && reconResp.StatusCode != http.StatusOK {
		log.Fatalf("Failed to create reconstruction job: %s", string(reconBody))
	}

	var reconJobResp struct {
		Data struct {
			JobID string `json:"job_id"`
		} `json:"data"`
	}
	json.Unmarshal(reconBody, &reconJobResp)
	reconJobID := reconJobResp.Data.JobID
	log.Printf("  Created job: %s (type: reconstruction)", reconJobID)

	// Print summary
	log.Println("")
	log.Println("=== Demo Seed Complete ===")
	log.Println("")
	log.Printf("Case ID:              %s", caseID)
	log.Printf("Scene Analysis Job:   %s", sceneJobID)
	log.Printf("Reconstruction Job:   %s", reconJobID)
	log.Println("")
	log.Println("Two AI jobs are now processing:")
	log.Println("  1. Scene Analysis - Gemini 3 Pro Vision detects objects and evidence")
	log.Println("  2. Reconstruction - HunyuanWorld-Mirror builds 3D room geometry")
	log.Println("")
	log.Printf("View the case at: http://localhost:3000/cases/%s", caseID)
	log.Println("")

	// Check job status via API
	log.Println("Checking job status...")
	for _, jid := range []string{sceneJobID, reconJobID} {
		checkURL := fmt.Sprintf("http://localhost:%s/v1/jobs/%s", cfg.Port, jid)
		checkResp, checkErr := http.Get(checkURL)
		if checkErr != nil {
			log.Printf("  %s: could not check status: %v", jid[:8], checkErr)
		} else {
			defer checkResp.Body.Close()
			checkBody, _ := io.ReadAll(checkResp.Body)
			var statusResp struct {
				Data struct {
					Status   string `json:"status"`
					Progress int    `json:"progress"`
					Type     string `json:"type"`
				} `json:"data"`
			}
			json.Unmarshal(checkBody, &statusResp)
			log.Printf("  %s (%s): %s (%d%%)", jid[:8], statusResp.Data.Type, statusResp.Data.Status, statusResp.Data.Progress)
		}
	}
}
