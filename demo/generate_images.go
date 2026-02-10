package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	geminiAPIKey = "AIzaSyDBn9ByLa2AIP_Gw7ddDNfu29YkoATVfn8"
	model        = "gemini-2.5-flash-image" // Nano Banana - image generation
	outputDir    = "images"
)

type GeminiRequest struct {
	Contents         []Content        `json:"contents"`
	GenerationConfig GenerationConfig `json:"generationConfig"`
}

type Content struct {
	Parts []Part `json:"parts"`
}

type Part struct {
	Text string `json:"text,omitempty"`
}

type GenerationConfig struct {
	ResponseMimeType  string   `json:"responseMimeType"`
	ResponseModalities []string `json:"responseModalities,omitempty"`
}

type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text       string `json:"text,omitempty"`
				InlineData *struct {
					MimeType string `json:"mimeType"`
					Data     string `json:"data"`
				} `json:"inlineData,omitempty"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func generateImage(prompt string, filename string) error {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, geminiAPIKey)

	reqBody := GeminiRequest{
		Contents: []Content{
			{
				Parts: []Part{
					{Text: prompt},
				},
			},
		},
		GenerationConfig: GenerationConfig{
			ResponseModalities: []string{"TEXT", "IMAGE"},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if geminiResp.Error != nil {
		return fmt.Errorf("API error: %s", geminiResp.Error.Message)
	}

	// Find image data in response
	for _, candidate := range geminiResp.Candidates {
		for _, part := range candidate.Content.Parts {
			if part.InlineData != nil && part.InlineData.Data != "" {
				// Decode base64 image
				imageData, err := base64.StdEncoding.DecodeString(part.InlineData.Data)
				if err != nil {
					return fmt.Errorf("failed to decode image: %w", err)
				}

				// Determine extension from mime type
				ext := ".png"
				if part.InlineData.MimeType == "image/jpeg" {
					ext = ".jpg"
				}

				// Save to file
				outputPath := filepath.Join(outputDir, filename+ext)
				if err := os.WriteFile(outputPath, imageData, 0644); err != nil {
					return fmt.Errorf("failed to save image: %w", err)
				}

				fmt.Printf("✓ Generated: %s\n", outputPath)
				return nil
			}
		}
	}

	return fmt.Errorf("no image found in response")
}

func main() {
	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("Failed to create output dir: %v\n", err)
		os.Exit(1)
	}

	// Define prompts for crime scene images
	prompts := []struct {
		name   string
		prompt string
	}{
		{
			name: "01_scene_overview",
			prompt: `Photorealistic crime scene photograph of a dimly lit office room at night.
The scene shows a ransacked executive office with an overturned desk chair, scattered papers on the floor,
open filing cabinet drawers, a desk lamp knocked over, and a broken window on the far wall with glass shards
on the floor. Yellow police evidence markers (numbered 1-5) are placed near key items.
Professional forensic photography style, wide angle shot, dramatic lighting from the broken window.
Do NOT include any people in the image.`,
		},
		{
			name: "02_entry_point",
			prompt: `Photorealistic close-up photograph of a broken office window at night - the entry point of a break-in.
The window frame shows clear signs of forced entry: shattered glass, bent metal frame, tool marks on the lock.
Glass shards scattered on the interior floor. A yellow evidence marker labeled "1" near the glass.
Moonlight streaming through the broken window. Forensic photography style with flash lighting.
Do NOT include any people in the image.`,
		},
		{
			name: "03_footprints",
			prompt: `Photorealistic forensic photograph of muddy boot footprints on an office carpet floor.
Multiple overlapping footprints showing a clear trail pattern from a window toward a desk area.
The boots appear to be work boots with distinctive tread patterns. Yellow evidence markers numbered 2 and 3.
A forensic measurement ruler placed next to one footprint. Professional crime scene photography lighting.
Do NOT include any people in the image.`,
		},
		{
			name: "04_desk_area",
			prompt: `Photorealistic crime scene photograph of a ransacked executive desk area.
The desk drawers are pulled open, papers scattered everywhere, a computer monitor pushed aside,
a coffee mug knocked over with dried coffee stain, desk lamp on its side still illuminated.
A single black glove dropped near the desk - evidence marker "4" next to it.
Signs of hurried search through documents. Dramatic shadows from the desk lamp.
Do NOT include any people in the image.`,
		},
		{
			name: "05_exit_door",
			prompt: `Photorealistic photograph of an office emergency exit door, slightly ajar.
The door shows signs of being used as an exit point: the push bar has smudges,
muddy footprints leading to and through the doorway, the door frame slightly damaged.
Yellow evidence marker "5" on the floor near the door. Red EXIT sign glowing above.
Dark hallway visible through the gap. Forensic photography style.
Do NOT include any people in the image.`,
		},
		{
			name: "06_evidence_closeup",
			prompt: `Photorealistic forensic close-up photograph of evidence items on an evidence table:
A single black leather glove, a small flashlight, and a crowbar tool - all in clear evidence bags
with white evidence labels. Professional forensic documentation photography with neutral background,
proper lighting for evidence documentation, measurement scale ruler visible.
Clean, clinical forensic lab setting.
Do NOT include any people in the image.`,
		},
	}

	fmt.Println("Generating crime scene images for SherlockOS demo...")
	fmt.Println("================================================")

	for i, p := range prompts {
		fmt.Printf("\n[%d/%d] Generating %s...\n", i+1, len(prompts), p.name)

		if err := generateImage(p.prompt, p.name); err != nil {
			fmt.Printf("✗ Error: %v\n", err)
		}

		// Rate limiting - wait between requests
		if i < len(prompts)-1 {
			time.Sleep(3 * time.Second)
		}
	}

	fmt.Println("\n================================================")
	fmt.Println("Image generation complete!")
}
