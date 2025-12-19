package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"net/http"

	"github.com/Brownie44l1/fer-api/internal/model"
	"github.com/nfnt/resize"
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

// preprocessImage converts an image to the format expected by the model
func (h *Handler) preprocessImage(img image.Image) ([]float32, error) {
	targetSize := uint(h.modelServer.Metadata.ImageSize)
	
	log.Printf("Target size from metadata: %d", targetSize)
	
	resized := resize.Resize(targetSize, targetSize, img, resize.Lanczos3)

	bounds := resized.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	
	log.Printf("Resized dimensions: %dx%d", width, height)
	
	channels := 3
	inputData := make([]float32, channels*width*height)
	
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, _ := resized.At(x, y).RGBA()
			
			rNorm := float32(r) / 65535.0
			gNorm := float32(g) / 65535.0
			bNorm := float32(b) / 65535.0
			
			pixelIndex := y*width + x
			inputData[pixelIndex] = rNorm
			inputData[width*height + pixelIndex] = gNorm
			inputData[2*width*height + pixelIndex] = bNorm
		}
	}

	log.Printf("Preprocessed image: %d values (3 channels × %d × %d)", len(inputData), width, height)

	return inputData, nil
}

func (h *Handler) PredictFromImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (10MB max)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "No image file provided. Use 'image' as the form field name", http.StatusBadRequest)
		return
	}
	defer file.Close()

	log.Printf("Received file: %s, size: %d bytes", header.Filename, header.Size)

	// Decode image
	img, format, err := image.Decode(file)
	if err != nil {
		http.Error(w, "Invalid image format. Supported: JPEG, PNG", http.StatusBadRequest)
		return
	}

	log.Printf("Image format: %s, dimensions: %dx%d", format, img.Bounds().Dx(), img.Bounds().Dy())

	// Preprocess image
	inputData, err := h.preprocessImage(img)
	if err != nil {
		log.Printf("Preprocessing error: %v", err)
		http.Error(w, "Failed to preprocess image", http.StatusInternalServerError)
		return
	}

	// Predict
	result, err := h.modelServer.Predict(inputData)
	if err != nil {
		log.Printf("Prediction error: %v", err)
		http.Error(w, "Prediction failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}