package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sherlockos/backend/internal/clients"
)

// PortraitHandler handles portrait generation via multi-turn chat
type PortraitHandler struct {
	imageClient *clients.GeminiImageGenClient
}

// NewPortraitHandler creates a new portrait handler
func NewPortraitHandler(imageClient *clients.GeminiImageGenClient) *PortraitHandler {
	return &PortraitHandler{imageClient: imageClient}
}

// PortraitChatRequest is the request body for portrait chat
type PortraitChatRequest struct {
	Messages []clients.PortraitChatMessage `json:"messages"`
}

// Chat handles POST /v1/portrait/chat
func (h *PortraitHandler) Chat(w http.ResponseWriter, r *http.Request) {
	if h.imageClient == nil {
		ServiceUnavailable(w, "Image generation not configured (GEMINI_API_KEY missing)")
		return
	}

	var req PortraitChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		BadRequest(w, "Invalid request body")
		return
	}

	if len(req.Messages) == 0 {
		BadRequest(w, "At least one message is required")
		return
	}

	// Validate messages
	for _, msg := range req.Messages {
		if msg.Role != "user" && msg.Role != "model" {
			BadRequest(w, fmt.Sprintf("Invalid role: %s (must be 'user' or 'model')", msg.Role))
			return
		}
	}

	// Call Gemini multi-turn image generation
	text, imageB64, err := h.imageClient.GeneratePortraitChat(r.Context(), req.Messages)
	if err != nil {
		InternalError(w, "Portrait generation failed: "+err.Error())
		return
	}

	Success(w, http.StatusOK, map[string]interface{}{
		"text":         text,
		"image_base64": imageB64,
	}, nil)
}
