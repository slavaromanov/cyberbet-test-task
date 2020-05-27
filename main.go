package main

import (
	"flag"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/reflection"

	pb "github.com/slavaromanov/cyberbet-test-task/proto"
	"github.com/slavaromanov/cyberbet-test-task/storage"
)

func main() {
	var (
		port, fileName string
		interval       time.Duration
	)
	flag.StringVar(&port, "port", "5330", "set gRPC server port")
	flag.StringVar(&fileName, "dump", "storage.gob", "file to save storage dump")
	flag.DurationVar(&interval, "interval", time.Minute, "interval for storage dump file")
	flag.Parse()

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		grpclog.Fatalf("failed to listen: %v", err)
	}
	store, err := storage.Open(fileName, interval)
	if err != nil {
		grpclog.Fatal(err)
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		store.Close()
		os.Exit(0)
	}()

	grpcServer := grpc.NewServer()
	pb.RegisterKeyValueStorageServer(grpcServer, &server{
		Storage: store,
	})
	reflection.Register(grpcServer)
	grpclog.Fatal(grpcServer.Serve(listener))
}
