package server

import (
	"context"
	"encoding/base64"
	"time"

	"github.com/google/uuid"
	"github.com/ryan-dayrit/pdf-extractor-pipeline/grpc-service/internal/kafka"
	"github.com/ryan-dayrit/pdf-extractor-pipeline/grpc-service/internal/store"
	pb "github.com/ryan-dayrit/pdf-extractor-pipeline/grpc-service/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCServer struct {
	pb.UnimplementedPDFExtractorServiceServer
	store    *store.Store
	producer *kafka.Producer
}

func NewGRPCServer(s *store.Store, p *kafka.Producer) *GRPCServer {
	return &GRPCServer{store: s, producer: p}
}

func (g *GRPCServer) UploadDocument(ctx context.Context, req *pb.UploadDocumentRequest) (*pb.UploadDocumentResponse, error) {
	id := uuid.NewString()
	doc := &store.Document{
		ID:        id,
		Filename:  req.Filename,
		Content:   req.Content,
		Status:    "uploaded",
		CreatedAt: time.Now(),
	}
	g.store.SaveDocument(doc)

	g.producer.Publish(map[string]interface{}{
		"event":            "document.uploaded",
		"document_id":      id,
		"filename":         req.Filename,
		"document_content": base64.StdEncoding.EncodeToString(req.Content),
	})

	return &pb.UploadDocumentResponse{DocumentId: id, Status: "uploaded"}, nil
}

func (g *GRPCServer) SubmitDataPoints(ctx context.Context, req *pb.SubmitDataPointsRequest) (*pb.SubmitDataPointsResponse, error) {
	doc, ok := g.store.GetDocument(req.DocumentId)
	if !ok {
		return nil, status.Errorf(codes.NotFound, "document not found")
	}

	dps := make([]store.DataPoint, len(req.DataPoints))
	dpsMaps := make([]map[string]string, len(req.DataPoints))
	for i, dp := range req.DataPoints {
		dps[i] = store.DataPoint{Name: dp.Name, Description: dp.Description}
		dpsMaps[i] = map[string]string{"name": dp.Name, "description": dp.Description}
	}
	g.store.SaveDataPoints(req.DocumentId, dps)

	g.producer.Publish(map[string]interface{}{
		"event":            "datapoints.submitted",
		"document_id":      req.DocumentId,
		"filename":         doc.Filename,
		"document_content": base64.StdEncoding.EncodeToString(doc.Content),
		"data_points":      dpsMaps,
	})

	return &pb.SubmitDataPointsResponse{Status: "processing"}, nil
}

func (g *GRPCServer) GetExtractionResults(ctx context.Context, req *pb.GetResultsRequest) (*pb.GetResultsResponse, error) {
	doc, ok := g.store.GetDocument(req.DocumentId)
	if !ok {
		return nil, status.Errorf(codes.NotFound, "document not found")
	}

	results := make([]*pb.ExtractionResult, len(doc.Results))
	for i, r := range doc.Results {
		results[i] = &pb.ExtractionResult{Name: r.Name, Value: r.Value, Confidence: r.Confidence}
	}
	return &pb.GetResultsResponse{
		DocumentId: doc.ID,
		Status:     doc.Status,
		Results:    results,
	}, nil
}

func (g *GRPCServer) ListDocuments(ctx context.Context, req *pb.ListDocumentsRequest) (*pb.ListDocumentsResponse, error) {
	docs := g.store.ListDocuments()
	pbDocs := make([]*pb.Document, len(docs))
	for i, d := range docs {
		pbDocs[i] = &pb.Document{
			Id:        d.ID,
			Filename:  d.Filename,
			Status:    d.Status,
			CreatedAt: d.CreatedAt.Format(time.RFC3339),
		}
	}
	return &pb.ListDocumentsResponse{Documents: pbDocs}, nil
}
