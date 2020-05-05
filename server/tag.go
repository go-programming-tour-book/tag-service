package server

import (
	"context"
	"encoding/json"

	"github.com/go-programming-tour-book/tag-service/pkg/bapi"
	"github.com/go-programming-tour-book/tag-service/pkg/errcode"

	//"encoding/json"

	//"github.com/go-programming-tour-book/tag-service/pkg/bapi"

	"google.golang.org/grpc"

	//"github.com/go-programming-tour-book/tag-service/pkg/errcode"
	pb "github.com/go-programming-tour-book/tag-service/proto"
)

type TagServer struct {
}

func NewTagServer() *TagServer {
	return &TagServer{}
}

func (t *TagServer) GetTagList(ctx context.Context, r *pb.GetTagListRequest) (*pb.GetTagListReply, error) {
	api := bapi.NewAPI("http://127.0.0.1:8000")
	body, err := api.GetTagList(ctx, r.GetName())
	if err != nil {
		return nil, errcode.TogRPCError(errcode.ErrorGetTagListFail)
	}

	tagList := pb.GetTagListReply{}
	err = json.Unmarshal(body, &tagList)
	if err != nil {
		return nil, errcode.TogRPCError(errcode.Fail)
	}
	return &tagList, nil
}

func GetClientConn(ctx context.Context, target string, opts []grpc.DialOption) (*grpc.ClientConn, error) {
	opts = append(opts, grpc.WithInsecure())
	return grpc.DialContext(ctx, target, opts...)
}
