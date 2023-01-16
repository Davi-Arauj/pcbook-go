package service

import (
	"errors"
	"fmt"
	"sync"

	"github.com/jinzhu/copier"
	"github.com/pcbook-go/pb"
)

var ErrAlreadyExists = errors.New("registro ja existe")

// LaptopStore é uma interface da loja do laptop
type LaptopStore interface {
	// Save salva um laptop na loja
	Save(laptop *pb.Laptop) error
	// Find busca um laptop pelo ID na loja
	Find(id string) (*pb.Laptop, error)
}

// InMemoryLaptopStore salva o laptop em memoria
type InMemoryLaptopStore struct {
	mutex sync.RWMutex
	data  map[string]*pb.Laptop
}

// NewInMemoryLaptopStore retorna um novo InMemoryLaptopStore
func NewInMemoryLaptopStore() *InMemoryLaptopStore {
	return &InMemoryLaptopStore{
		data: make(map[string]*pb.Laptop),
	}
}

// Save salva o laptop para a loja
func (store *InMemoryLaptopStore) Save(laptop *pb.Laptop) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if store.data[laptop.Id] != nil {
		return ErrAlreadyExists
	}

	// copia profunda
	other := &pb.Laptop{}
	err := copier.Copy(other, laptop)
	if err != nil {
		return fmt.Errorf("não foi possivel copiar os dados do laptop: %w", err)
	}

	store.data[other.Id] = other
	return nil
}

// Find busca um laptop pelo ID
func (store *InMemoryLaptopStore) Find(id string) (*pb.Laptop, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	laptop := store.data[id]
	if laptop == nil {
		return nil, nil
	}

	// copia profunda
	other := &pb.Laptop{}
	err := copier.Copy(other, laptop)
	if err != nil {
		return nil, fmt.Errorf("não foi possivel copiar os dados do laptop: %w", err)
	}

	return other, nil
}
