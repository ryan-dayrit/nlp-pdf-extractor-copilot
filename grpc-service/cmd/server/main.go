package main

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/ryan-dayrit/pdf-extractor-pipeline/grpc-service/internal/kafka"
	"github.com/ryan-dayrit/pdf-extractor-pipeline/grpc-service/internal/server"
	"github.com/ryan-dayrit/pdf-extractor-pipeline/grpc-service/internal/store"
	pb "github.com/ryan-dayrit/pdf-extractor-pipeline/grpc-service/proto"
	"google.golang.org/grpc"
)

var (
	globalStore    *store.Store
	globalProducer *kafka.Producer
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func jsonResponse(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func handleUploadDocument(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid multipart form"})
		return
	}
	f, header, err := r.FormFile("file")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "missing file field"})
		return
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": "read error"})
		return
	}

	id := uuid.NewString()
	doc := &store.Document{
		ID:        id,
		Filename:  header.Filename,
		Content:   content,
		Status:    "uploaded",
		CreatedAt: time.Now(),
	}
	globalStore.SaveDocument(doc)

	globalProducer.Publish(map[string]interface{}{
		"event":            "document.uploaded",
		"document_id":      id,
		"filename":         header.Filename,
		"document_content": base64.StdEncoding.EncodeToString(content),
	})

	jsonResponse(w, http.StatusOK, map[string]string{"document_id": id, "status": "uploaded"})
}

func handleSubmitDataPoints(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	doc, ok := globalStore.GetDocument(id)
	if !ok {
		jsonResponse(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}

	var body struct {
		DataPoints []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"data_points"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	dps := make([]store.DataPoint, len(body.DataPoints))
	dpsMaps := make([]map[string]string, len(body.DataPoints))
	for i, dp := range body.DataPoints {
		dps[i] = store.DataPoint{Name: dp.Name, Description: dp.Description}
		dpsMaps[i] = map[string]string{"name": dp.Name, "description": dp.Description}
	}
	globalStore.SaveDataPoints(id, dps)

	globalProducer.Publish(map[string]interface{}{
		"event":            "datapoints.submitted",
		"document_id":      id,
		"filename":         doc.Filename,
		"document_content": base64.StdEncoding.EncodeToString(doc.Content),
		"data_points":      dpsMaps,
	})

	jsonResponse(w, http.StatusOK, map[string]string{"status": "processing"})
}

func handleGetResults(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	doc, ok := globalStore.GetDocument(id)
	if !ok {
		jsonResponse(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}

	results := make([]map[string]interface{}, len(doc.Results))
	for i, res := range doc.Results {
		results[i] = map[string]interface{}{
			"name":       res.Name,
			"value":      res.Value,
			"confidence": res.Confidence,
		}
	}
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"document_id": doc.ID,
		"status":      doc.Status,
		"results":     results,
	})
}

func handleListDocuments(w http.ResponseWriter, r *http.Request) {
	docs := globalStore.ListDocuments()
	result := make([]map[string]string, len(docs))
	for i, d := range docs {
		result[i] = map[string]string{
			"id":         d.ID,
			"filename":   d.Filename,
			"status":     d.Status,
			"created_at": d.CreatedAt.Format(time.RFC3339),
		}
	}
	jsonResponse(w, http.StatusOK, map[string]interface{}{"documents": result})
}

func handlePutResults(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var body struct {
		Results []struct {
			Name       string  `json:"name"`
			Value      string  `json:"value"`
			Confidence float64 `json:"confidence"`
		} `json:"results"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	results := make([]store.ExtractionResult, len(body.Results))
	for i, r := range body.Results {
		results[i] = store.ExtractionResult{Name: r.Name, Value: r.Value, Confidence: r.Confidence}
	}
	status := body.Status
	if status == "" {
		status = "completed"
	}
	if !globalStore.SaveResults(id, results, status) {
		jsonResponse(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	jsonResponse(w, http.StatusOK, map[string]string{"status": "ok"})
}

func startHTTPServer() {
	r := mux.NewRouter()
	r.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		jsonResponse(w, http.StatusOK, map[string]string{"status": "ok"})
	}).Methods(http.MethodGet)
	r.HandleFunc("/api/documents", handleUploadDocument).Methods(http.MethodPost)
	r.HandleFunc("/api/documents", handleListDocuments).Methods(http.MethodGet)
	r.HandleFunc("/api/documents/{id}/datapoints", handleSubmitDataPoints).Methods(http.MethodPost)
	r.HandleFunc("/api/documents/{id}/results", handleGetResults).Methods(http.MethodGet)
	r.HandleFunc("/api/documents/{id}/results", handlePutResults).Methods(http.MethodPut)

	// OPTIONS for CORS preflight
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.WriteHeader(http.StatusNoContent)
		}
	})

	log.Println("HTTP server listening on :8080")
	if err := http.ListenAndServe(":8080", corsMiddleware(r)); err != nil {
		log.Fatalf("HTTP server error: %v", err)
	}
}

func main() {
	globalStore = store.New()

	var err error
	// Retry Kafka connection
	for i := 0; i < 10; i++ {
		globalProducer, err = kafka.NewProducer()
		if err == nil {
			break
		}
		log.Printf("Kafka not ready, retrying in 3s... (%v)", err)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		log.Printf("WARNING: Kafka producer unavailable: %v. Continuing without Kafka.", err)
		globalProducer = kafka.NewNoopProducer()
	}

	// Start HTTP server in background
	go startHTTPServer()

	// Start gRPC server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	svc := server.NewGRPCServer(globalStore, globalProducer)
	pb.RegisterPDFExtractorServiceServer(grpcServer, svc)

	log.Println("gRPC server listening on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("gRPC server error: %v", err)
	}
}
