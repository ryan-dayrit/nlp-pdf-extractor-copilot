package nlp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type DataPoint struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ExtractionResult struct {
	Name       string  `json:"name"`
	Value      string  `json:"value"`
	Confidence float64 `json:"confidence"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient() *Client {
	url := os.Getenv("NLP_SERVICE_URL")
	if url == "" {
		url = "http://localhost:8000"
	}
	return &Client{
		baseURL:    url,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *Client) ExtractDataPoints(documentContent string, dataPoints []DataPoint) ([]ExtractionResult, error) {
	payload := map[string]interface{}{
		"document_content": documentContent,
		"data_points":      dataPoints,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Post(c.baseURL+"/extract-datapoints", "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("nlp request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("nlp service returned status %d", resp.StatusCode)
	}

	var result struct {
		Results []ExtractionResult `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode nlp response: %w", err)
	}
	return result.Results, nil
}
