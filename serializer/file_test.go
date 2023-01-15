package serializer_test

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/pcbook-go/pb"
	"github.com/pcbook-go/sample"
	"github.com/pcbook-go/serializer"
	"github.com/stretchr/testify/require"
)

func TestFileSerializer(t *testing.T) {
	t.Parallel()

	binaryFile := "../tmp/laptop.bin"
	jsonFile := "../tmp/laptop.json"

	// criando um laptop1
	laptop1 := sample.NewLaptop()

	// transformando o laptop1 em binario
	err := serializer.WriteProtobufToBinaryFile(laptop1, binaryFile)
	require.NoError(t, err)

	// transformando o laptop1 em json
	err = serializer.WriteProtobufToJSONFile(laptop1, jsonFile)
	require.NoError(t, err)

	// lendo o laptop do arquivo binario e gravando em um laptop2
	laptop2 := &pb.Laptop{}
	err = serializer.ReadProtobufFromBinaryFile(binaryFile, laptop2)
	require.NoError(t, err)

	// comparando o laptop1 com o laptop2
	require.True(t, proto.Equal(laptop1, laptop2))
}
