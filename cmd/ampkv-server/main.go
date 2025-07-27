package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Unfield/AmpKV/drivers/cache/ristretto"
	"github.com/Unfield/AmpKV/drivers/store/badger"
	"github.com/Unfield/AmpKV/internal/auth"
	"github.com/Unfield/AmpKV/internal/logger"
	"github.com/Unfield/AmpKV/internal/server"
	pb "github.com/Unfield/AmpKV/pkg/client/rpc"
	"github.com/Unfield/AmpKV/pkg/embedded"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	var (
		grpcPort = flag.Int("grpc-port", 50051, "The gRPC server port")
		dbPath   = flag.String("db-path", "ampkv_server.db", "Path to the AmpKV Embedded DB file")
		httpMode = flag.String("http-mode", "http", "Http/Https mode for the Http Server")
	)
	flag.Parse()

	logger.InitLogger()
	defer logger.GetLogger().Sync()

	appLogger := logger.GetLogger()

	appLogger.Info("Starting AmpKV...")

	ampkvCache, err := ristretto.NewRistrettoCache(1e7, 1<<30, 64)
	if err != nil {
		appLogger.Fatal("failed to initialize ristretto cache")
	}

	if dbPath == nil {
		appLogger.Fatal("db-path must not be empty")
	}

	ampkvStore, err := badger.NewBadgerStore(*dbPath)
	if err != nil {
		appLogger.Fatal("failed to initialize badger store")
	}

	ampkvEmbedded, err := embedded.NewAmpKV(ampkvCache, ampkvStore, embedded.AmpKVOptions{DefaultTTL: 10 * time.Minute})
	if err != nil {
		appLogger.Fatal("failed to initialize AmpKV embedded")
	}
	defer func() {
		err := ampkvEmbedded.Close()
		if err != nil {
			appLogger.Error("failed to close AmpKV embedded", zap.Error(err))
		} else {
			appLogger.Info("AmpKV embedded closed successfully")
		}
	}()

	grpcServerImpl := server.NewAmpKVGrpcServer(ampkvEmbedded)

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", *grpcPort))
	if err != nil {
		appLogger.Fatal("Failed to listen", zap.Error(err))
	}

	apiKeyManager, err := auth.NewApiKeyManager(ampkvEmbedded)
	if err != nil {
		appLogger.Fatal("Failed to initialize api key manager", zap.Error(err))
	}

	s := grpc.NewServer(grpc.UnaryInterceptor(server.AuthUnaryServerInterceptor(apiKeyManager)))
	pb.RegisterAmpKVServiceServer(s, grpcServerImpl)

	reflection.Register(s)

	go func() {
		appLogger.Info("gRPC server running", zap.String("address", lis.Addr().String()))
		if err := s.Serve(lis); err != nil {
			appLogger.Fatal("gRPC server failed to server", zap.Error(err))
		}
	}()

	httpServerImpl := server.NewAmpKVHttpServer(ampkvEmbedded, apiKeyManager)

	if *httpMode == "https" {
		httpServerImpl.ListenAutoTLS("0.0.0.0", 4443)
	} else {
		httpServerImpl.Listen("0.0.0.0", 8080)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	appLogger.Info("Shutting down", zap.String("signal", sig.String()))
	s.GracefulStop()
	appLogger.Info("gRPC server stopped")
	appLogger.Info("AmpKV server exited")
}
