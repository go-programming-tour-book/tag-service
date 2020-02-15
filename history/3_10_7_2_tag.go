package history

import (
	"context"

	"github.com/go-programming-tour-book/tag-service/internal/middleware"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"

	"google.golang.org/grpc"

	"github.com/go-programming-tour-book/tag-service/pkg/errcode"
	pb "github.com/go-programming-tour-book/tag-service/proto"
)

type TagServer struct {
}

func NewTagServer() *TagServer {
	return &TagServer{}
}

func (t *TagServer) GetTagList(ctx context.Context, r *pb.GetTagListRequest) (*pb.GetTagListReply, error) {
	clientConn, err := GetClientConn(ctx, "localhost:8004", []grpc.DialOption{grpc.WithUnaryInterceptor(
		grpc_middleware.ChainUnaryClient(
			middleware.UnaryContextTimeout(),
			middleware.ClientTracing(),
		),
	)})
	if err != nil {
		return nil, errcode.TogRPCError(errcode.Fail)
	}
	defer clientConn.Close()
	tagServiceClient := pb.NewTagServiceClient(clientConn)
	resp, err := tagServiceClient.GetTagList(ctx, &pb.GetTagListRequest{Name: "Go"})
	if err != nil {
		return nil, errcode.TogRPCError(errcode.Fail)
	}

	return resp, nil
}

func GetClientConn(ctx context.Context, target string, opts []grpc.DialOption) (*grpc.ClientConn, error) {
	opts = append(opts, grpc.WithInsecure())
	return grpc.DialContext(ctx, target, opts...)
}
