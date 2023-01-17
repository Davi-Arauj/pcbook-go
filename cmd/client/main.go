package main

import (
	"bufio"
	"context"
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/pcbook-go/pb"
	"github.com/pcbook-go/sample"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func createLaptop(laptopClient pb.LaptopServiceClient, laptop *pb.Laptop) {
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
			log.Print("laptop já existe")
		} else {
			log.Fatal("não pode criar laptop: ", err)
		}
		return
	}

	log.Printf("laptop criado com id: %s", res.Id)
}

func searchLaptop(laptopClient pb.LaptopServiceClient, filter *pb.Filter) {
	log.Print("filtro de pesquisa: ", filter)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.SearchLaptopRequest{Filter: filter}
	stream, err := laptopClient.SearchLaptop(ctx, req)
	if err != nil {
		log.Fatal("erro ao buscar laptop: ", err)
	}

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Fatal("erro ao receber resposta: ", err)
		}

		laptop := res.GetLaptop()
		log.Print("- found: ", laptop.GetId())
		log.Print("  + brand: ", laptop.GetBrand())
		log.Print("  + name: ", laptop.GetName())
		log.Print("  + cpu cores: ", laptop.GetCpu().GetNumberCores())
		log.Print("  + cpu min ghz: ", laptop.GetCpu().GetMinGhz())
		log.Print("  + ram: ", laptop.GetRam())
		log.Print("  + price: ", laptop.GetPriceUsd())

	}
}

func uploadImage(laptopClient pb.LaptopServiceClient, laptopID string, imagePath string) {
	file, err := os.Open(imagePath)
	if err != nil {
		log.Fatal("erro ao abrir arquivo da imagem: ", err)
	}
	defer file.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := laptopClient.UploadImage(ctx)
	if err != nil {
		log.Fatal("erro ao carregar imagem: ", err)

	}

	req := &pb.UploadImageRequest{
		Data: &pb.UploadImageRequest_Info{
			Info: &pb.ImageInfo{
				LaptopId:  laptopID,
				ImageType: filepath.Ext(imagePath),
			},
		},
	}

	err = stream.Send(req)
	if err != nil {
		log.Fatal("erro ao enviar imagem para o servidor: ", err, stream.RecvMsg(nil))
	}
	
	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatal("erro ao ler parte do buffer: ", err)
		}

		req := &pb.UploadImageRequest{
			Data: &pb.UploadImageRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		}

		err = stream.Send(req)
		if err != nil {
			log.Fatal("erro ao enviar parte da imagem ao servidor: ", err, stream.RecvMsg(nil))
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal("erro ao receber resposta: ", err)
	}

	log.Printf("imagem carregada com o id: %s, de tamanhao: %d", res.GetId(), res.GetSize())
}
func testCreateLaptop(laptopClient pb.LaptopServiceClient) {
	createLaptop(laptopClient, sample.NewLaptop())
}

func testSearchLaptop(laptopClient pb.LaptopServiceClient) {
	for i := 0; i < 10; i++ {
		createLaptop(laptopClient, sample.NewLaptop())
	}

	filter := &pb.Filter{
		MaxPriceUsd: 3000,
		MinCpuCores: 4,
		MinCpuGhz:   2.5,
		MinRam:      &pb.Memory{Value: 8, Unit: pb.Memory_GIGABYTE},
	}

	searchLaptop(laptopClient, filter)
}

func testUploadImage(laptopClient pb.LaptopServiceClient) {
	laptop := sample.NewLaptop()
	createLaptop(laptopClient, laptop)
	uploadImage(laptopClient, laptop.GetId(), "tmp/laptop.jpg")
}

func main() {
	serverAddress := flag.String("address", "", "o endereço do servidor")
	flag.Parse()
	log.Printf("servidor de discagem %s", *serverAddress)

	conn, err := grpc.Dial(*serverAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatal("não pode discar para o servidor: ", err)
	}

	laptopClient := pb.NewLaptopServiceClient(conn)
	testUploadImage(laptopClient)
}
