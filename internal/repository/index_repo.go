package repository

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"sync"

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
	UpdateIndex(idx IdentifiableIndex) error
}

type Snapshotter interface {
	SnapshotExport() error
	SnapshotImport() error
	IsSnapshotEnabled() bool
	ToggleSnapshots()
}

type LocalIndexRepo struct {
	SnapshotDir     string
	SnapshotEnabled bool
	indexes         sync.Map
	toSnapshot      sync.Map
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
	r.indexes.Store(idx.ID, idx)
	r.markIndexForSnapshot(idx.ID)
	return nil
}

func (r *LocalIndexRepo) DeleteIndex(id string) error {
	r.toSnapshot.Delete(id)
	_, ok := r.indexes.LoadAndDelete(id)
	if !ok {
		return ErrIndexDoesNotExist
	}
	if r.IsSnapshotEnabled() {
		// because of route handler and validations, should never reach path traversal error
		out, err := util.SafeJoin(r.SnapshotDir, id+indexExtension)
		if err != nil {
			return err
		}
		return os.Remove(out)
	}
	return nil
}

func (r *LocalIndexRepo) markIndexForSnapshot(id string) error {
	if r.IsSnapshotEnabled() {
		log.Println("Marked index for snapshot:", id)
		r.toSnapshot.Store(id, struct{}{})
	}
	return nil
}

func (r *LocalIndexRepo) UpdateIndex(idx IdentifiableIndex) error {
	return r.markIndexForSnapshot(idx.ID)
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
		// check if index is marked for snapshot and unmark it
		if _, ok := r.toSnapshot.LoadAndDelete(idx.ID); !ok {
			continue
		}
		log.Println("Snapshotting index:", idx.ID)
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
			r.indexes.Store(id, IdentifiableIndex{ID: id, Index: idx})
		}
	}
	return nil
}

func (r *LocalIndexRepo) IsSnapshotEnabled() bool {
	return r.SnapshotEnabled
}

func (r *LocalIndexRepo) ToggleSnapshots() {
	r.SnapshotEnabled = !r.SnapshotEnabled
}
