package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// nlpRequest is the body sent to the NLP service.
type nlpRequest struct {
	PDFB64     string   `json:"pdf_base64"`
	DataPoints []string `json:"data_points"`
}

// nlpResponse is the response from the NLP service.
type nlpResponse struct {
	Results map[string]string `json:"results"`
}

// callNLPService posts the PDF data and data points to the NLP service,
// retrying up to 3 times on failure.
func callNLPService(pdfBase64 string, dataPoints []string) (map[string]string, error) {
	nlpServiceURL := getEnv("NLP_SERVICE_URL", "http://nlp-service:8000")
	url := fmt.Sprintf("%s/extract", nlpServiceURL)

	reqBody, err := json.Marshal(nlpRequest{
		PDFB64:     pdfBase64,
		DataPoints: dataPoints,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal NLP request: %w", err)
	}

	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		log.Printf("Calling NLP service (attempt %d/3): %s", attempt, url)

		resp, err := http.Post(url, "application/json", bytes.NewReader(reqBody)) //nolint:noctx
		if err != nil {
			lastErr = fmt.Errorf("attempt %d: POST to NLP service: %w", attempt, err)
			log.Printf("NLP service call failed: %v", lastErr)
			if attempt < 3 {
				time.Sleep(time.Duration(attempt) * time.Second)
			}
			continue
		}

		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()

		if readErr != nil {
			lastErr = fmt.Errorf("attempt %d: read NLP response body: %w", attempt, readErr)
			log.Printf("NLP response read failed: %v", lastErr)
			if attempt < 3 {
				time.Sleep(time.Duration(attempt) * time.Second)
			}
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			lastErr = fmt.Errorf("attempt %d: NLP service returned status %d: %s", attempt, resp.StatusCode, body)
			log.Printf("NLP service error: %v", lastErr)
			if attempt < 3 {
				time.Sleep(time.Duration(attempt) * time.Second)
			}
			continue
		}

		var nlpResp nlpResponse
		if err := json.Unmarshal(body, &nlpResp); err != nil {
			lastErr = fmt.Errorf("attempt %d: unmarshal NLP response: %w", attempt, err)
			log.Printf("NLP response unmarshal failed: %v", lastErr)
			if attempt < 3 {
				time.Sleep(time.Duration(attempt) * time.Second)
			}
			continue
		}

		log.Printf("NLP service returned %d results", len(nlpResp.Results))
		return nlpResp.Results, nil
	}

	return nil, fmt.Errorf("NLP service failed after 3 attempts: %w", lastErr)
}
