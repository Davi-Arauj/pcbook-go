package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/pcbook-go/pb"
	"github.com/pcbook-go/sample"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func main() {
	serverAddress := flag.String("endereço", "", "o endereço do servidor")
	flag.Parse()
	log.Printf("servidor de discagem %s", *serverAddress)

	conn, err := grpc.Dial(*serverAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatal("não pode discar para o servidor: ", err)
	}

	laptopClient := pb.NewLaptopServiceClient(conn)

	laptop := sample.NewLaptop()
	laptop.Id = ""
	req := &pb.CreateLaptopRequest{
		Laptop: laptop,
	}

	// definir tempo limite
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := laptopClient.CreateLaptop(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.AlreadyExists {
			// não é grande coisa
			log.Print("laptop já existe")
		} else {
			log.Fatal("não pode criar laptop: ", err)
		}
		return
	}

	log.Printf("laptop criado com id: %s", res.Id)
}
