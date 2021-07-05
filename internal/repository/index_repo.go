package repository

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/ruial/busca/internal/util"
	"github.com/ruial/busca/pkg/index"
)

const indexExtension = ".out"

var (
	ErrIndexAlreadyExists = errors.New("Index already exists")
	ErrIndexDoesNotExist  = errors.New("Index does not exist")
	ErrInvalidIndex       = errors.New("Index ID has invalid characters")
)

type IdentifiableIndex struct {
	ID    string
	Index *index.Index
}

func NewIdentifiableIndex(id string, index *index.Index) (idx IdentifiableIndex, err error) {
	// do not allow slashes and dots on index names to prevent file traversal vulnerabilities
	if strings.ContainsRune(id, '/') ||
		strings.ContainsRune(id, '\\') ||
		strings.ContainsRune(id, '.') ||
		strings.TrimSpace(id) == "" {
		return idx, ErrInvalidIndex
	}
	idx.ID = id
	idx.Index = index
	return
}

type IndexRepo interface {
	GetIndexes() []IdentifiableIndex
	GetIndex(id string) (IdentifiableIndex, bool)
	CreateIndex(idx IdentifiableIndex) error
	DeleteIndex(id string) error
	SnapshotExport() error
	SnapshotImport() error
	IsSnapshotEnabled() bool
}

type LocalIndexRepo struct {
	SnapshotDir      string
	SnapshotInterval time.Duration
	indexes          sync.Map
}

func (r *LocalIndexRepo) GetIndexes() (list []IdentifiableIndex) {
	r.indexes.Range(func(key, value interface{}) bool {
		list = append(list, value.(IdentifiableIndex))
		return true
	})
	return
}

func (r *LocalIndexRepo) GetIndex(id string) (idx IdentifiableIndex, ok bool) {
	value, ok := r.indexes.Load(id)
	if !ok {
		return idx, ok
	}
	return value.(IdentifiableIndex), ok
}

func (r *LocalIndexRepo) CreateIndex(idx IdentifiableIndex) error {
	if _, ok := r.indexes.Load(idx.ID); ok {
		return ErrIndexAlreadyExists
	}
	r.indexes.Store(string(idx.ID), idx)
	return nil
}

func (r *LocalIndexRepo) DeleteIndex(id string) error {
	_, ok := r.indexes.LoadAndDelete(id)
	if !ok {
		return ErrIndexDoesNotExist
	}
	if r.IsSnapshotEnabled() {
		out, err := util.SafeJoin(r.SnapshotDir, id+indexExtension)
		if err != nil {
			return err
		}
		return os.Remove(out)
	}
	return nil
}

func (r *LocalIndexRepo) SnapshotExport() error {
	if !r.IsSnapshotEnabled() {
		return nil
	}
	if _, err := os.Stat(r.SnapshotDir); os.IsNotExist(err) {
		if err := os.Mkdir(r.SnapshotDir, 0700); err != nil {
			return fmt.Errorf("Error creating export directory: %s", err.Error())
		}
	}
	for _, idx := range r.GetIndexes() {
		out, err := util.SafeJoin(r.SnapshotDir, idx.ID+indexExtension)
		if err != nil {
			return err
		}
		// cloning is faster than serializing, so lock time is reduced for readers
		if err := index.Export(idx.Index.Clone(), out); err != nil {
			return fmt.Errorf("Error exporting index %s: %s", idx.ID, err.Error())
		}
	}
	return nil
}

func (r *LocalIndexRepo) SnapshotImport() error {
	if r.SnapshotDir == "" {
		return nil
	}
	files, err := os.ReadDir(r.SnapshotDir)
	if err != nil {
		return fmt.Errorf("Error reading directory: %s", err.Error())
	}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), indexExtension) && !file.IsDir() {
			id := strings.TrimSuffix(file.Name(), indexExtension)
			idx, err := index.Import(path.Join(r.SnapshotDir, file.Name()))
			if err != nil {
				return fmt.Errorf("Error importing index %s: %s", id, err.Error())
			}
			r.CreateIndex(IdentifiableIndex{ID: id, Index: idx})
		}
	}
	return nil
}

func (r *LocalIndexRepo) IsSnapshotEnabled() bool {
	return r.SnapshotInterval > 0
}
