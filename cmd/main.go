package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"path"
	"strings"

	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/go-programming-tour-book/tag-service/global"
	"github.com/go-programming-tour-book/tag-service/internal/middleware"
	"github.com/go-programming-tour-book/tag-service/pkg/swagger"
	"github.com/go-programming-tour-book/tag-service/pkg/tracer"
	pb "github.com/go-programming-tour-book/tag-service/proto"
	"github.com/go-programming-tour-book/tag-service/server"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var port string

func init() {
	flag.StringVar(&port, "port", "8004", "启动端口号")
	flag.Parse()

	err := setupTracer()
	if err != nil {
		log.Fatalf("init.setupTracer err: %v", err)
	}
}

func setupTracer() error {
	jaegerTracer, _, err := tracer.NewJaegerTracer("tour-service", "127.0.0.1:6831")
	if err != nil {
		return err
	}
	global.Tracer = jaegerTracer
	return nil
}

func main() {
	err := RunServer(port)
	if err != nil {
		log.Fatalf("Run Serve err: %v", err)
	}
}

func runGrpcGatewayServer() *gwruntime.ServeMux {
	endpoint := "0.0.0.0:" + port
	gwmux := gwruntime.NewServeMux()
	dopts := []grpc.DialOption{grpc.WithInsecure()}
	_ = pb.RegisterTagServiceHandlerFromEndpoint(context.Background(), gwmux, endpoint, dopts)

	return gwmux
}

func RunServer(port string) error {
	httpMux := runHttpServer()
	grpcS := runGrpcServer()
	gatewayMux := runGrpcGatewayServer()

	httpMux.Handle("/", gatewayMux)
	return http.ListenAndServe(":"+port, grpcHandlerFunc(grpcS, httpMux))
}

func runHttpServer() *http.ServeMux {
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`pong`))
	})
	prefix := "/swagger-ui/"
	fileServer := http.FileServer(&assetfs.AssetFS{
		Asset:    swagger.Asset,
		AssetDir: swagger.AssetDir,
		Prefix:   "third_party/swagger-ui",
	})

	serveMux.Handle(prefix, http.StripPrefix(prefix, fileServer))
	serveMux.HandleFunc("/swagger/", func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "swagger.json") {
			http.NotFound(w, r)
			return
		}

		p := strings.TrimPrefix(r.URL.Path, "/swagger/")
		p = path.Join("proto", p)

		http.ServeFile(w, r, p)
	})

	return serveMux
}

func runGrpcServer() *grpc.Server {
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			middleware.AccessLog,
			middleware.ErrorLog,
			middleware.Recovery,
			middleware.ServerTracing,
		)),
	}
	s := grpc.NewServer(opts...)
	pb.RegisterTagServiceServer(s, server.NewTagServer())
	reflection.Register(s)

	return s
}

func grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	}), &http2.Server{})
}
