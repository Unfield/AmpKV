package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Unfield/AmpKV/drivers/cache/ristretto"
	"github.com/Unfield/AmpKV/drivers/store/badger"
	"github.com/Unfield/AmpKV/internal/server"
	pb "github.com/Unfield/AmpKV/pkg/client/rpc"
	"github.com/Unfield/AmpKV/pkg/embedded"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	var (
		grpcPort = flag.Int("grpc-port", 50051, "The gRPC server port")
		dbPath   = flag.String("db-path", "ampkv_server.db", "Path to the AmpKV Embedded DB file")
	)
	flag.Parse()

	ampkvCache, err := ristretto.NewRistrettoCache(1e7, 1<<30, 64)
	if err != nil {
		log.Fatal("failed to initialize ristretto cache")
	}

	if dbPath == nil {
		log.Fatal("db-path must not be empty")
	}

	ampkvStore, err := badger.NewBadgerStore(*dbPath)
	if err != nil {
		log.Fatal("failed to initialize badger store")
	}

	ampkvEmbedded, err := embedded.NewAmpKV(ampkvCache, ampkvStore, embedded.AmpKVOptions{DefaultTTL: 10 * time.Minute})
	if err != nil {
		log.Fatal("failed to initialize AmpKV embedded")
	}
	defer func() {
		err := ampkvEmbedded.Close()
		if err != nil {
			log.Printf("failed to close AmpKV embedded: %v", err)
		} else {
			log.Println("AmpKV embedded closed successfully")
		}
	}()

	grpcServerImpl := server.NewAmpKVGrpcServer(*ampkvEmbedded)

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", *grpcPort))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterAmpKVServiceServer(s, grpcServerImpl)

	reflection.Register(s)

	go func() {
		log.Printf("gRPC server running on %s", lis.Addr())
		if err := s.Serve(lis); err != nil {
			log.Fatalf("gRPC server failed to serve: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	log.Printf("Recived signal %s, gracefully shutting down...", sig)

	s.GracefulStop()
	log.Println("gRPC server gracefully stopped")
	log.Println("AmpKV server exited")
}
