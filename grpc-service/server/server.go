// Package server implements both the gRPC ExtractorService and a plain
// HTTP/JSON REST server on port 8080 (for the Svelte frontend).
package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ryan-dayrit/nlp-pdf-extractor/grpc-service/kafka"
	"github.com/ryan-dayrit/nlp-pdf-extractor/grpc-service/pb"
)

// Document is the in-memory representation of an uploaded PDF.
type Document struct {
	ID         string
	Filename   string
	Status     string
	DataPoints []string
	Results    map[string]string
	PDFData    []byte
}

// Server holds the in-memory store, the Kafka producer, and serves both gRPC
// and HTTP traffic.
type Server struct {
	mu       sync.RWMutex
	docs     map[string]*Document
	producer *kafka.Producer
}

// NewServer constructs a Server. producer may be nil if Kafka is unavailable.
func NewServer(producer *kafka.Producer) *Server {
	return &Server{
		docs:     make(map[string]*Document),
		producer: producer,
	}
}

// ---------------------------------------------------------------------------
// gRPC service implementation
// ---------------------------------------------------------------------------

func (s *Server) UploadDocument(_ context.Context, req *pb.UploadDocumentRequest) (*pb.UploadDocumentResponse, error) {
	id := uuid.New().String()
	doc := &Document{
		ID:         id,
		Filename:   req.Filename,
		Status:     "pending",
		DataPoints: req.DataPoints,
		PDFData:    req.PdfData,
		Results:    make(map[string]string),
	}

	s.mu.Lock()
	s.docs[id] = doc
	s.mu.Unlock()

	if s.producer != nil {
		pdfBase64 := base64.StdEncoding.EncodeToString(req.PdfData)
		if err := s.producer.PublishDocumentUpload(id, req.Filename, pdfBase64, req.DataPoints); err != nil {
			log.Printf("warning: kafka publish failed: %v", err)
		}
	}

	return &pb.UploadDocumentResponse{DocumentId: id, Status: "pending"}, nil
}

func (s *Server) GetDataPoints(_ context.Context, req *pb.GetDataPointsRequest) (*pb.GetDataPointsResponse, error) {
	s.mu.RLock()
	doc, ok := s.docs[req.DocumentId]
	s.mu.RUnlock()

	if !ok {
		return nil, status.Errorf(codes.NotFound, "document %s not found", req.DocumentId)
	}

	results := make(map[string]string, len(doc.Results))
	for k, v := range doc.Results {
		results[k] = v
	}
	return &pb.GetDataPointsResponse{
		DocumentId: doc.ID,
		Status:     doc.Status,
		Results:    results,
	}, nil
}

func (s *Server) ListDocuments(_ context.Context, _ *pb.ListDocumentsRequest) (*pb.ListDocumentsResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	summaries := make([]*pb.DocumentSummary, 0, len(s.docs))
	for _, doc := range s.docs {
		summaries = append(summaries, &pb.DocumentSummary{
			DocumentId: doc.ID,
			Filename:   doc.Filename,
			Status:     doc.Status,
		})
	}
	return &pb.ListDocumentsResponse{Documents: summaries}, nil
}

func (s *Server) UpdateDataPoints(_ context.Context, req *pb.UpdateDataPointsRequest) (*pb.UpdateDataPointsResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	doc, ok := s.docs[req.DocumentId]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "document %s not found", req.DocumentId)
	}

	for k, v := range req.Results {
		doc.Results[k] = v
	}
	doc.Status = "completed"

	return &pb.UpdateDataPointsResponse{Status: "updated"}, nil
}

// ---------------------------------------------------------------------------
// HTTP REST server
// ---------------------------------------------------------------------------

// NewHTTPMux builds a net/http ServeMux with all REST routes and CORS middleware.
func (s *Server) NewHTTPMux() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /documents", s.handleUploadDocument)
	mux.HandleFunc("GET /documents", s.handleListDocuments)
	mux.HandleFunc("GET /documents/{id}/datapoints", s.handleGetDataPoints)
	mux.HandleFunc("POST /documents/{id}/datapoints", s.handleUpdateDataPoints)
	return corsMiddleware(mux)
}

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

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writeJSON encode error: %v", err)
	}
}

// POST /documents — multipart form: field "file" (PDF), field "data_points" (JSON array string)
func (s *Server) handleUploadDocument(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "failed to parse multipart form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "field 'file' is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	pdfData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "failed to read file", http.StatusInternalServerError)
		return
	}

	var dataPoints []string
	if dpJSON := r.FormValue("data_points"); dpJSON != "" {
		if err := json.Unmarshal([]byte(dpJSON), &dataPoints); err != nil {
			// fall back: treat the raw value as a single data point
			dataPoints = []string{dpJSON}
		}
	}

	resp, err := s.UploadDocument(r.Context(), &pb.UploadDocumentRequest{
		Filename:   header.Filename,
		PdfData:    pdfData,
		DataPoints: dataPoints,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

// GET /documents — list all documents
func (s *Server) handleListDocuments(w http.ResponseWriter, r *http.Request) {
	resp, _ := s.ListDocuments(r.Context(), &pb.ListDocumentsRequest{})
	writeJSON(w, http.StatusOK, resp)
}

// GET /documents/{id}/datapoints — fetch extraction results for a document
func (s *Server) handleGetDataPoints(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	resp, err := s.GetDataPoints(r.Context(), &pb.GetDataPointsRequest{DocumentId: id})
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// POST /documents/{id}/datapoints — called by the NLP consumer to store results
// Body: {"results": {"key": "value", ...}}
func (s *Server) handleUpdateDataPoints(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var body struct {
		Results map[string]string `json:"results"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	resp, err := s.UpdateDataPoints(r.Context(), &pb.UpdateDataPointsRequest{
		DocumentId: id,
		Results:    body.Results,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}
