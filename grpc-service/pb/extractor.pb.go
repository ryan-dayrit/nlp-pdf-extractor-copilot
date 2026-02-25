// Hand-written protobuf message types (mirrors extractor.proto).
// These plain Go structs are encoded via the JSON codec registered in extractor_grpc.pb.go.
package pb

// UploadDocumentRequest is the request for UploadDocument.
type UploadDocumentRequest struct {
	Filename   string   `json:"filename"`
	PdfData    []byte   `json:"pdf_data"`
	DataPoints []string `json:"data_points"`
}

// UploadDocumentResponse is the response from UploadDocument.
type UploadDocumentResponse struct {
	DocumentId string `json:"document_id"`
	Status     string `json:"status"`
}

// GetDataPointsRequest is the request for GetDataPoints.
type GetDataPointsRequest struct {
	DocumentId string `json:"document_id"`
}

// GetDataPointsResponse is the response from GetDataPoints.
type GetDataPointsResponse struct {
	DocumentId string            `json:"document_id"`
	Status     string            `json:"status"`
	Results    map[string]string `json:"results"`
}

// ListDocumentsRequest is the request for ListDocuments.
type ListDocumentsRequest struct{}

// ListDocumentsResponse is the response from ListDocuments.
type ListDocumentsResponse struct {
	Documents []*DocumentSummary `json:"documents"`
}

// DocumentSummary is a brief representation of a stored document.
type DocumentSummary struct {
	DocumentId string `json:"document_id"`
	Filename   string `json:"filename"`
	Status     string `json:"status"`
}

// UpdateDataPointsRequest is the request for UpdateDataPoints.
type UpdateDataPointsRequest struct {
	DocumentId string            `json:"document_id"`
	Results    map[string]string `json:"results"`
}

// UpdateDataPointsResponse is the response from UpdateDataPoints.
type UpdateDataPointsResponse struct {
	Status string `json:"status"`
}
