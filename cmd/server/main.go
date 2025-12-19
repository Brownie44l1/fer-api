package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Brownie44l1/fer-api/internal/handlers"
	"github.com/Brownie44l1/fer-api/internal/model"
)

func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next(w, r)
	}
}

func main() {
	// Get the project root directory
	execPath, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}
	
	// If running from cmd/server, go up two levels
	if filepath.Base(execPath) == "server" {
		execPath = filepath.Join(execPath, "../..")
	}

	modelPath := filepath.Join(execPath, "models", "model_embedded.onnx")
	metadataPath := filepath.Join(execPath, "models", "model_metadata.json")

	log.Printf("Loading model from: %s", modelPath)

	modelServer, err := model.NewServer(modelPath, metadataPath)
	if err != nil {
		log.Fatalf("Failed to initialize model server: %v", err)
	}
	defer modelServer.Close()

	handler := handlers.NewHandler(modelServer)

    http.HandleFunc("/health", enableCORS(handler.Health))
	http.HandleFunc("/predict", enableCORS(handler.Predict))
    http.HandleFunc("/predict/image", enableCORS(handler.PredictFromImage))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Printf("Model loaded: %s", modelPath)
	log.Printf("Classes: %v", modelServer.Metadata.Classes)
	log.Println("Endpoints:")
	log.Println("  GET /health - Health check")
        log.Println("  POST /predict - Raw array prediction")
        log.Println("  POST /predict/image   - Predict from image upload")
	log.Printf("\n💡 Upload test: curl -X POST -F \"image=@face.jpg\" http://localhost:%s/predict/image\n\n", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}