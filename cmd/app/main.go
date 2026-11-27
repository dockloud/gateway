package main

import (
	"context"
	"log"
	"net/http"
	"time"

	authv1 "github.com/dockloud/contract/gen/auth/v1"
	docv1 "github.com/dockloud/contract/gen/document/v1"
	linkv1 "github.com/dockloud/contract/gen/link/v1"
	model "github.com/dockloud/gateway/cmd/internal/domain/config"
	"github.com/dockloud/gateway/cmd/internal/infrastructure/config"
	"github.com/dockloud/service-auth/core"
	"github.com/dockloud/service-auth/grpcauth"
	"github.com/dockloud/user-auth/pkg/infrastructure/verifier"
	"github.com/dockloud/user-auth/pkg/transport/http/metadata"
	"github.com/dockloud/user-auth/pkg/transport/http/middleware"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func establishConnection(conf *model.MergedConfig, name string, authInterceptor grpc.UnaryClientInterceptor) *grpc.ClientConn {
	depConfig, ok := conf.Dependencies[name]
	if !ok {
		log.Fatalf("dependency %s not found", name)
	}

	conn, err := grpc.NewClient(
		depConfig.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(authInterceptor),
	)
	if err != nil {
		log.Fatal(err)
	}

	return conn
}

func main() {
	ctx := context.Background()
	conf, err := config.LoadMergedConfig()
	if err != nil {
		log.Fatal(err)
	}

	key := serviceauth.Key{
		ID:      conf.ID,
		Service: conf.Name,
		Secret:  conf.JWT.ServiceSecret,
	}
	signer, err := serviceauth.NewSigner(key)
	if err != nil {
		log.Fatal(err)
	}
	authInterceptor, err := grpcauth.NewClientInterceptor(signer, 0)
	if err != nil {
		log.Fatal(err)
	}

	authConn := establishConnection(conf, "auth", authInterceptor)
	defer authConn.Close()
	docConn := establishConnection(conf, "document", authInterceptor)
	defer docConn.Close()
	linkConn := establishConnection(conf, "link", authInterceptor)
	defer linkConn.Close()

	mux := runtime.NewServeMux(
		runtime.WithMetadata(metadata.MetadataWithUID),
	)

	if err := authv1.RegisterAuthServiceHandler(ctx, mux, authConn); err != nil {
		log.Fatal(err)
	}
	if err := docv1.RegisterDocumentServiceHandler(ctx, mux, docConn); err != nil {
		log.Fatal(err)
	}
	if err := linkv1.RegisterLinkServiceHandler(ctx, mux, linkConn); err != nil {
		log.Fatal(err)
	}

	ver, err := verifier.NewVerifier(conf.JWT.UserAuthToken)
	if err != nil {
		log.Fatal(err)
	}

	handler := middleware.UserAuthMiddleware(
		ver,
		conf.PublicRoutes...,
	)(mux)
	server := &http.Server{
		Addr:              conf.Port,
		Handler:           handler,
		ReadHeaderTimeout: time.Duration(conf.Timeout) * time.Second,
	}

	log.Printf("HTTP gateway listening on %s", conf.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
