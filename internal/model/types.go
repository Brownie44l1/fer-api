package model

type Metadata struct {
	InputShape  []int64  `json:"input_shape"`
	OutputShape []int64  `json:"output_shape"`
	Classes     []string `json:"classes"`
	ImageSize   int      `json:"image_size"`
}

type PredictionRequest struct {
	Image []float32 `json:"image"`
}

type PredictionResponse struct {
	Class       string             `json:"class"`
	Confidence  float32            `json:"confidence"`
	Predictions map[string]float32 `json:"predictions"`
}