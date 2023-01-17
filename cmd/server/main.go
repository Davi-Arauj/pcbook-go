package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/pcbook-go/pb"
	"github.com/pcbook-go/service"
	"google.golang.org/grpc"
)

func main() {
	port := flag.Int("port", 0, "a porta do servidor")
	flag.Parse()
	log.Printf("o servidor está na porta %d", *port)

	laptopStore := service.NewInMemoryLaptopStore()
	imageStore := service.NewDiskIMageStore("img")
	laptopServer := service.NewLaptopServer(laptopStore, imageStore)
	grpcServer := grpc.NewServer()
	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)

	addres := fmt.Sprintf("0.0.0.0:%d", *port)
	listener, err := net.Listen("tcp", addres)
	if err != nil {
		log.Fatal("não foi possivel inicar o servidor", err)
	}

	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("não foi possivel inicar o servidor", err)
	}
}
