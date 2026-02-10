package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	supabaseURL    = "https://hdfaugwofzqqdjuzcsin.supabase.co"
	supabaseSecret = "sb_secret_3MTp1hBYNrR6egqO4RqMvQ_2oAi-uc7"
	bucket         = "assets"
	caseID         = "36aaae7d-9c52-44f7-a820-449631001ea8"
	batchID        = "d31b9a07-10e8-4e78-a932-e32b6e25590e"
)

func uploadImage(filepath string, filename string) error {
	// Read file
	data, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Build storage key
	storageKey := fmt.Sprintf("cases/%s/scans/%s/%s", caseID, batchID, filename)
	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", supabaseURL, bucket, storageKey)

	// Create request
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("apikey", supabaseSecret)
	req.Header.Set("Authorization", "Bearer "+supabaseSecret)
	req.Header.Set("Content-Type", "image/png")
	req.Header.Set("x-upsert", "true") // Overwrite if exists

	// Execute request
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func main() {
	imageDir := "images"

	files, err := os.ReadDir(imageDir)
	if err != nil {
		fmt.Printf("Failed to read images directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Uploading images to Supabase Storage...")
	fmt.Printf("Bucket: %s\n", bucket)
	fmt.Printf("Case ID: %s\n", caseID)
	fmt.Println("")

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".png" {
			continue
		}

		filePath := filepath.Join(imageDir, file.Name())
		fmt.Printf("Uploading %s... ", file.Name())

		if err := uploadImage(filePath, file.Name()); err != nil {
			fmt.Printf("FAILED: %v\n", err)
		} else {
			fmt.Println("OK")
		}
	}

	fmt.Println("\nUpload complete!")
}
