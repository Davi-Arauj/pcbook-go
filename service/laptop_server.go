package service

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/pcbook-go/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LaptopServer é um servidor que provê serviços do laptop
type LaptopServer struct {
	pb.UnimplementedLaptopServiceServer
	Store LaptopStore
}

// NewLaptopServer retorna um novo LaptopServer
func NewLaptopServer(store LaptopStore) *LaptopServer {
	return &LaptopServer{Store: store}
}

// CreateLaptop é um RPC unario para criar um novo Laptop
func (server *LaptopServer) CreateLaptop(
	ctx context.Context,
	req *pb.CreateLaptopRequest,
) (*pb.CreateLaptopResponse, error) {
	laptop := req.GetLaptop()
	log.Printf("uma solicitação de criação de um laptop foi recebido: %s", laptop.Id)

	if len(laptop.Id) > 0 {
		// verificando se o id é valido
		_, err := uuid.Parse(laptop.Id)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "o ID do laptop não é valido UUID: %v", err)
		}
	} else {
		id, err := uuid.NewRandom()
		if err != nil {
			return nil, status.Errorf(codes.Internal, "erro ao gerar um novo laptop ID: %v", err)
		}
		laptop.Id = id.String()
	}

	// algum processamento pesado
	time.Sleep(6 * time.Second)

	if ctx.Err() == context.Canceled {
		log.Print("solicitação cancelada")
		return nil, status.Error(codes.Canceled, "solicitação cancelada")
	}

	if ctx.Err() == context.DeadlineExceeded {
		log.Print("o tempo foi excedido")
		return nil, status.Error(codes.DeadlineExceeded, "o tempo foi excedido")
	}

	// salvando o laptop na loja
	err := server.Store.Save(laptop)
	if err != nil {
		code := codes.Internal
		if errors.Is(err, ErrAlreadyExists) {
			code = codes.AlreadyExists
		}

		return nil, status.Errorf(code, "não foi possivel salvar o laptop na loja: %v", err)
	}

	log.Printf("o laptop foi salvo de id: %v", laptop.Id)

	res := &pb.CreateLaptopResponse{
		Id: laptop.Id,
	}

	return res, nil
}

// SearchLaptop é um RPC de streaming de servidor para procurar laptops
func (server *LaptopServer) SearchLaptop(
	req *pb.SearchLaptopRequest,
	stream pb.LaptopService_SearchLaptopServer,
) (outErr error) {
	filter := req.GetFilter()
	log.Printf("receber uma solicitação de pesquisa de laptop com filtro: %v", filter)

	err := server.Store.Search(
		stream.Context(),
		filter,
		func(laptop *pb.Laptop) {
			res := &pb.SearchLaptopResponse{Laptop: laptop}
			err := stream.Send(res)
			if err != nil {
				outErr = status.Errorf(codes.Unknown, "não pode enviar resposta: %v", err)
				return
			}

			log.Printf("enviou laptop com id: %s", laptop.GetId())
		},
	)

	if err != nil {
		return status.Errorf(codes.Internal, "erro inesperado: %v", err)
	}

	return nil
}