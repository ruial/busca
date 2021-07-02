package repository

import (
	"errors"
	"sync"

	"github.com/ruial/busca/pkg/index"
)

var (
	ErrIndexAlreadyExists = errors.New("Index already exists")
	ErrIndexDoesNotExist  = errors.New("Index does not exist")
)

type IdentifiableIndex struct {
	ID    string
	Index *index.Index
}

type IndexRepo interface {
	GetIndexes() []IdentifiableIndex
	GetIndex(id string) (IdentifiableIndex, bool)
	AddIndex(idx IdentifiableIndex) error
	DeleteIndex(id string) error
}

type InMemoryIndexRepo struct {
	indexes *sync.Map
}

func NewInMemoryIndexRepo() InMemoryIndexRepo {
	return InMemoryIndexRepo{indexes: &sync.Map{}}
}

func (r *InMemoryIndexRepo) GetIndexes() (list []IdentifiableIndex) {
	r.indexes.Range(func(key, value interface{}) bool {
		list = append(list, value.(IdentifiableIndex))
		return true
	})
	return
}

func (r *InMemoryIndexRepo) GetIndex(id string) (idx IdentifiableIndex, ok bool) {
	value, ok := r.indexes.Load(id)
	if !ok {
		return idx, ok
	}
	return value.(IdentifiableIndex), ok
}

func (r *InMemoryIndexRepo) CreateIndex(idx IdentifiableIndex) error {
	if _, ok := r.indexes.Load(idx.ID); ok {
		return ErrIndexAlreadyExists
	}
	r.indexes.Store(string(idx.ID), idx)
	return nil
}

func (r *InMemoryIndexRepo) DeleteIndex(id string) error {
	_, ok := r.indexes.LoadAndDelete(id)
	if !ok {
		return ErrIndexDoesNotExist
	}
	return nil
}
