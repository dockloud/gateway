package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	authv1 "github.com/dockloud/contract/gen/auth/v1"
	doc1 "github.com/dockloud/contract/gen/document/v1"
	"github.com/dockloud/user-auth/pkg/infrastructure/verifier"
	"github.com/dockloud/user-auth/pkg/transport/http/metadata"
	"github.com/dockloud/user-auth/pkg/transport/http/middleware"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func establishConnection(addr string) *grpc.ClientConn {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal(err)
	}

	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(conn)

	return conn
}

func main() {
	ctx := context.Background()

	authConn := establishConnection(os.Getenv("AUTH_CONN_ADDR"))
	docConn := establishConnection(os.Getenv("DOC_CONN_ADDR"))

	mux := runtime.NewServeMux(
		runtime.WithMetadata(metadata.MetadataWithUID),
	)

	if err := authv1.RegisterAuthServiceHandler(ctx, mux, authConn); err != nil {
		log.Fatal(err)
	}
	if err := doc1.RegisterDocumentServiceHandler(ctx, mux, docConn); err != nil {
		log.Fatal(err)
	}

	ver, err := verifier.NewVerifier(os.Getenv("JWT_AUTH_KEY"))
	if err != nil {
		log.Fatal(err)
	}

	handler := middleware.UserAuthMiddleware(ver)(mux)
	server := &http.Server{
		Addr:              ":8080",
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("HTTP gateway listening on :8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
