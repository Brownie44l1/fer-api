package model

import (
	"encoding/json"
	"fmt"
	"os"

	ort "github.com/yalue/onnxruntime_go"
)

type Server struct {
	session      *ort.AdvancedSession
	Metadata     Metadata
	inputTensor  *ort.Tensor[float32]
	outputTensor *ort.Tensor[float32]
}

func NewServer(modelPath, metadataPath string) (*Server, error) {
	if err := ort.InitializeEnvironment(); err != nil {
		return nil, fmt.Errorf("failed to initialize ONNX environment: %w", err)
	}

	metaFile, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	var metadata Metadata
	if err := json.Unmarshal(metaFile, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	inputShape := ort.NewShape(metadata.InputShape...)
	outputShape := ort.NewShape(metadata.OutputShape...)

	inputTensor, err := ort.NewEmptyTensor[float32](inputShape)
	if err != nil {
		return nil, fmt.Errorf("failed to create input tensor: %w", err)
	}

	outputTensor, err := ort.NewEmptyTensor[float32](outputShape)
	if err != nil {
		inputTensor.Destroy()
		return nil, fmt.Errorf("failed to create output tensor: %w", err)
	}

	session, err := ort.NewAdvancedSession(modelPath,
		[]string{"input"}, []string{"output"},
		[]ort.ArbitraryTensor{inputTensor}, []ort.ArbitraryTensor{outputTensor},
		nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create ONNX session: %w", err)
	}

	return &Server{
		session:      session,
		Metadata:     metadata,
		inputTensor:  inputTensor,
		outputTensor: outputTensor,
	}, nil
}

func (s *Server) Predict(inputData []float32) (*PredictionResponse, error) {
	copy(s.inputTensor.GetData(), inputData)

	if err := s.session.Run(); err != nil {
		return nil, fmt.Errorf("inference failed: %w", err)
	}

	outputData := s.outputTensor.GetData()

	maxIdx := 0
	maxVal := outputData[0]
	predictions := make(map[string]float32)

	for i, val := range outputData {
		if i < len(s.Metadata.Classes) {
			predictions[s.Metadata.Classes[i]] = val
			if val > maxVal {
				maxVal = val
				maxIdx = i
			}
		}
	}

	return &PredictionResponse{
		Class:       s.Metadata.Classes[maxIdx],
		Confidence:  maxVal,
		Predictions: predictions,
	}, nil
}

func (s *Server) Close() {
	if s.inputTensor != nil {
		s.inputTensor.Destroy()
	}
	if s.outputTensor != nil {
		s.outputTensor.Destroy()
	}
	if s.session != nil {
		s.session.Destroy()
	}
	ort.DestroyEnvironment()
}