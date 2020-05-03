package middleware

import (
	"context"
	"time"

	"github.com/go-programming-tour-book/tag-service/global"
	"github.com/go-programming-tour-book/tag-service/pkg/metatext"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"google.golang.org/grpc/metadata"

	"google.golang.org/grpc"
)

// ClientInterceptor grpc client wrapper
func ClientTracing() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		var parentCtx opentracing.SpanContext
		var spanOpts []opentracing.StartSpanOption
		var parentSpan = opentracing.SpanFromContext(ctx)
		if parentSpan != nil {
			parentCtx = parentSpan.Context()
			spanOpts = append(spanOpts, opentracing.ChildOf(parentCtx))
		}
		spanOpts = append(spanOpts, []opentracing.StartSpanOption{
			opentracing.Tag{Key: string(ext.Component), Value: "gRPC"},
			ext.SpanKindRPCClient,
		}...)

		span := global.Tracer.StartSpan(method, spanOpts...)
		defer span.Finish()

		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}
		_ = global.Tracer.Inject(span.Context(), opentracing.TextMap, metatext.MetadataTextMap{md})
		newCtx := opentracing.ContextWithSpan(metadata.NewOutgoingContext(ctx, md), span)
		return invoker(newCtx, method, req, reply, cc, opts...)
	}
}

func UnaryContextTimeout() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, resp interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx, cancel := defaultContextTimeout(ctx)
		if cancel != nil {
			defer cancel()
		}

		return invoker(ctx, method, req, resp, cc, opts...)
	}
}

func StreamContextTimeout() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		ctx, cancel := defaultContextTimeout(ctx)
		if cancel != nil {
			defer cancel()
		}

		return streamer(ctx, desc, cc, method, opts...)
	}
}

func defaultContextTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		defaultTimeout := 60 * time.Second
		ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
	}

	return ctx, cancel
}
