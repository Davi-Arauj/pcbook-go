package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"

	"github.com/google/uuid"
	"github.com/pcbook-go/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LaptopServer é um servidor que provê serviços do laptop
type LaptopServer struct {
	pb.UnimplementedLaptopServiceServer
	laptopStore LaptopStore
	imageStore  ImageStore
}

// NewLaptopServer retorna um novo LaptopServer
func NewLaptopServer(laptopStore LaptopStore, imageStore ImageStore) *LaptopServer {
	return &LaptopServer{
		UnimplementedLaptopServiceServer: pb.UnimplementedLaptopServiceServer{},
		laptopStore:                      laptopStore,
		imageStore:                       imageStore,
	}
}

const maxImageSize = 1 << 20

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
	//time.Sleep(6 * time.Second)

	if err := contextError(ctx); err != nil {
		return nil, err
	}

	// salvando o laptop na loja
	err := server.laptopStore.Save(laptop)
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
) error {
	filter := req.GetFilter()
	log.Printf("receber uma solicitação de pesquisa de laptop com filtro: %v", filter)

	err := server.laptopStore.Search(
		stream.Context(),
		filter,
		func(laptop *pb.Laptop) error {
			res := &pb.SearchLaptopResponse{Laptop: laptop}
			err := stream.Send(res)
			if err != nil {
				return err
			}

			log.Printf("enviou laptop com id: %s", laptop.GetId())
			return nil
		},
	)

	if err != nil {
		return status.Errorf(codes.Internal, "erro inesperado: %v", err)
	}

	return nil
}

func (server *LaptopServer) UploadImage(stream pb.LaptopService_UploadImageServer) error {
	req, err := stream.Recv()
	if err != nil {
		return logError(status.Errorf(codes.Unknown, "erro ao receber informações da imagem"))
	}

	laptopID := req.GetInfo().GetLaptopId()
	imageType := req.GetInfo().GetImageType()

	log.Printf("solicitação de upload de imagem: %s para o laptop: %s", imageType, laptopID)

	laptop, err := server.laptopStore.Find(laptopID)
	if err != nil {
		return logError(status.Errorf(codes.Internal, "erro ao buscar o laptop: %v", err))
	}
	if laptop == nil {
		return logError(status.Errorf(codes.InvalidArgument, "laptop id %s não existe", laptopID))
	}

	imageData := bytes.Buffer{}
	imageSize := 0

	for {
		err := contextError(stream.Context())
		if err != nil {
			return err
		}
		log.Print("aguardando para receber mais dados")

		req, err := stream.Recv()
		if err == io.EOF {
			log.Print("não há mais dados")
			break
		}
		if err != nil {
			return logError(status.Errorf(codes.Unknown, "erro ao receber partes dos dados: %v", err))
		}

		chunk := req.GetChunkData()
		size := len(chunk)

		log.Printf("recebendo uma parte do total: %d", size)

		imageSize += size
		if imageSize > maxImageSize {
			return logError(status.Errorf(codes.InvalidArgument, "a imagem é muito grande: %d > %d", imageSize, maxImageSize))
		}
		_, err = imageData.Write(chunk)
		if err != nil {
			return logError(status.Errorf(codes.Internal, "erro ao gravar parte do dado: %v", err))
		}
	}

	imageID, err := server.imageStore.Save(laptopID, imageType, imageData)
	if err != nil {
		return logError(status.Errorf(codes.Internal, "erro ao salvar a imagem: %v", err))
	}

	res := &pb.UploadImageResponse{
		Id:   imageID,
		Size: uint32(imageSize),
	}

	err = stream.SendAndClose(res)
	if err != nil {
		return logError(status.Errorf(codes.Unknown, "erro ao enviar resposta: %v", err))
	}

	log.Printf("imagem salva com o id: %s, size: %d", imageID, imageSize)
	return nil
}

func contextError(ctx context.Context) error {
	switch ctx.Err() {
	case context.Canceled:
		return logError(status.Error(codes.Canceled, "requisição cancelada"))
	case context.DeadlineExceeded:
		return logError(status.Error(codes.DeadlineExceeded, "o tempo foi esgotado"))
	default:
		return nil
	}
}

func logError(err error) error {
	if err != nil {
		log.Print(err)
	}
	return err
}
