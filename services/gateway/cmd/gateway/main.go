package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	satellitev1 "github.com/satellite-communications/gateway/gen/satellite/v1"
	"github.com/satellite-communications/gateway/internal/config"
	"github.com/satellite-communications/gateway/internal/events"
	"github.com/satellite-communications/gateway/internal/httpserver"
	"github.com/satellite-communications/gateway/internal/server"
	"google.golang.org/grpc"
)

func main() {
	cfg, err := config.Load()

	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)

	defer stop()

	broadcaster := events.NewBroadcaster()
	gatewayServer := server.NewGatewayServer(broadcaster)

	httpServer := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           httpserver.NewHandler(gatewayServer, broadcaster),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	failures := make(chan error, 1)

	grpcServer, err := startGRPCServer(cfg.GRPCPort, ctx, gatewayServer)

	if err != nil {
		log.Fatal(err)
	}

	go func() {
		log.Printf("gateway HTTP listening on %s", httpServer.Addr)
		fmt.Println()
		failures <- httpServer.ListenAndServe()
	}()

	select {
	//failures branch still needs coordinated cleanup.
	case err := <-failures:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err) //calls os.Exit(1) immediately, meaning deferred cleanup functions don't run.
		}
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancel()

		grpcStopped := make(chan struct{})

		go func() {
			grpcServer.GracefulStop()
			close(grpcStopped)
		}()

		if err := httpServer.Shutdown(shutdown); err != nil {
			log.Printf("shutdown: %v", err)
			_ = httpServer.Close()
		}

		select {
		case <-grpcStopped:
			log.Println("gRPC server stopped gracefully")
		case <-shutdown.Done():
			log.Println("gRPC shutdown timed out; forcing stop")
			grpcServer.Stop()
		}

	}
}

/*
logs errors from Serve() rather than forwarding them to main error channel.
*/
func startGRPCServer(port string, ctx context.Context, gatewayServer *server.GatewayServer) (*grpc.Server, error) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return nil, err
	}

	grpcServer := grpc.NewServer()

	go gatewayServer.MonitorSatellites(ctx)
	go gatewayServer.MonitorCommunications(ctx)

	satellitev1.RegisterSatelliteGatewayServiceServer(
		grpcServer,
		gatewayServer,
	)

	log.Println("gRPC server listening on port 50051")

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Printf("gRPC server stopped: %v", err)
		}
	}()

	return grpcServer, nil
}
