package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/Brownie44l1/fer-api/internal/model"
)

type Handler struct {
	modelServer *model.Server
}

func NewHandler(modelServer *model.Server) *Handler {
	return &Handler{
		modelServer: modelServer,
	}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func (h *Handler) Predict(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	var req model.PredictionRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	expectedSize := int(h.modelServer.Metadata.InputShape[0])
	for _, dim := range h.modelServer.Metadata.InputShape[1:] {
		expectedSize *= int(dim)
	}

	if len(req.Image) != expectedSize {
		http.Error(w, fmt.Sprintf("Expected %d values, got %d", expectedSize, len(req.Image)),
			http.StatusBadRequest)
		return
	}

	result, err := h.modelServer.Predict(req.Image)
	if err != nil {
		log.Printf("Prediction error: %v", err)
		http.Error(w, "Prediction failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}