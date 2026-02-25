// Hand-written gRPC service descriptors for ExtractorService.
// Registers a JSON codec named "proto" so that plain Go structs (no protoc) are
// wire-compatible between our internal services.
package pb

import (
	"context"
	"encoding/json"

	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
)

func init() {
	// Replace the default protobuf codec with JSON so we can use plain Go structs.
	encoding.RegisterCodec(jsonCodec{})
}

// jsonCodec serialises gRPC messages as JSON.
type jsonCodec struct{}

func (jsonCodec) Name() string                        { return "proto" }
func (jsonCodec) Marshal(v interface{}) ([]byte, error) { return json.Marshal(v) }
func (jsonCodec) Unmarshal(b []byte, v interface{}) error { return json.Unmarshal(b, v) }

// ExtractorServiceServer is the server-side interface for ExtractorService.
type ExtractorServiceServer interface {
	UploadDocument(context.Context, *UploadDocumentRequest) (*UploadDocumentResponse, error)
	GetDataPoints(context.Context, *GetDataPointsRequest) (*GetDataPointsResponse, error)
	ListDocuments(context.Context, *ListDocumentsRequest) (*ListDocumentsResponse, error)
	UpdateDataPoints(context.Context, *UpdateDataPointsRequest) (*UpdateDataPointsResponse, error)
}

// UnimplementedExtractorServiceServer provides default (stub) implementations.
type UnimplementedExtractorServiceServer struct{}

func (UnimplementedExtractorServiceServer) UploadDocument(_ context.Context, _ *UploadDocumentRequest) (*UploadDocumentResponse, error) {
	return nil, nil
}
func (UnimplementedExtractorServiceServer) GetDataPoints(_ context.Context, _ *GetDataPointsRequest) (*GetDataPointsResponse, error) {
	return nil, nil
}
func (UnimplementedExtractorServiceServer) ListDocuments(_ context.Context, _ *ListDocumentsRequest) (*ListDocumentsResponse, error) {
	return nil, nil
}
func (UnimplementedExtractorServiceServer) UpdateDataPoints(_ context.Context, _ *UpdateDataPointsRequest) (*UpdateDataPointsResponse, error) {
	return nil, nil
}

// RegisterExtractorServiceServer registers srv with the given gRPC server.
func RegisterExtractorServiceServer(s *grpc.Server, srv ExtractorServiceServer) {
	s.RegisterService(&ExtractorService_ServiceDesc, srv)
}

// --- method handlers ---------------------------------------------------------

func _UploadDocument_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UploadDocumentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExtractorServiceServer).UploadDocument(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/extractor.ExtractorService/UploadDocument"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExtractorServiceServer).UploadDocument(ctx, req.(*UploadDocumentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GetDataPoints_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetDataPointsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExtractorServiceServer).GetDataPoints(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/extractor.ExtractorService/GetDataPoints"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExtractorServiceServer).GetDataPoints(ctx, req.(*GetDataPointsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ListDocuments_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListDocumentsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExtractorServiceServer).ListDocuments(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/extractor.ExtractorService/ListDocuments"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExtractorServiceServer).ListDocuments(ctx, req.(*ListDocumentsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UpdateDataPoints_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateDataPointsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExtractorServiceServer).UpdateDataPoints(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/extractor.ExtractorService/UpdateDataPoints"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExtractorServiceServer).UpdateDataPoints(ctx, req.(*UpdateDataPointsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ExtractorService_ServiceDesc is the grpc.ServiceDesc for ExtractorService.
var ExtractorService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "extractor.ExtractorService",
	HandlerType: (*ExtractorServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "UploadDocument", Handler: _UploadDocument_Handler},
		{MethodName: "GetDataPoints", Handler: _GetDataPoints_Handler},
		{MethodName: "ListDocuments", Handler: _ListDocuments_Handler},
		{MethodName: "UpdateDataPoints", Handler: _UpdateDataPoints_Handler},
	},
	Streams: []grpc.StreamDesc{},
}
