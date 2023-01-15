package serializer

import (
	"fmt"
	"io/ioutil"

	"github.com/golang/protobuf/proto"
)

// WriteProtobufToJSONFile grava a mensagem do buffer de protocolo no arquivo JSON
func WriteProtobufToJSONFile(message proto.Message, filename string) error {
	data, err := ProtobufToJSON(message)
	if err != nil {
		return fmt.Errorf("não é possível empacotar a mensagem proto para JSON: %w", err)
	}

	err = ioutil.WriteFile(filename, []byte(data), 0644)
	if err != nil {
		return fmt.Errorf("não é possível gravar dados JSON no arquivo: %w", err)
	}

	return nil
}

// WriteProtobufToBinaryFile grava a mensagem do buffer de protocolo no arquivo binário
func WriteProtobufToBinaryFile(message proto.Message, filename string) error {
	data, err := proto.Marshal(message)
	if err != nil {
		return fmt.Errorf("não é possível empacotar mensagem proto para binário: %w", err)
	}

	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("não é possível gravar dados binários no arquivo: %w", err)
	}

	return nil
}

// ReadProtobufFromBinaryFile lê a mensagem do buffer de protocolo do arquivo binário
func ReadProtobufFromBinaryFile(filename string, message proto.Message) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("não é possível ler dados binários do arquivo: %w", err)
	}

	err = proto.Unmarshal(data, message)
	if err != nil {
		return fmt.Errorf("não é possível descompactar o binário para a mensagem proto: %w", err)
	}

	return nil
}