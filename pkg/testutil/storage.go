package testutil

import (
	"bytes"
	"context"
	"io"
	"sync"

	"go/kir-tube/internal/media"
)

// FakeStorage is an in-memory media.ObjectStorage for tests: it keeps objects in
// a map instead of talking to MinIO, so media flows can be exercised without any
// external infrastructure.
type FakeStorage struct {
	mu      sync.RWMutex
	objects map[string]fakeObject
}

type fakeObject struct {
	data        []byte
	contentType string
}

// NewFakeStorage returns an empty in-memory store.
func NewFakeStorage() *FakeStorage {
	return &FakeStorage{objects: make(map[string]fakeObject)}
}

// compile-time check that FakeStorage satisfies the media contract.
var _ media.ObjectStorage = (*FakeStorage)(nil)

func (s *FakeStorage) Put(_ context.Context, key string, r io.Reader, _ int64, contentType string) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.objects[key] = fakeObject{data: data, contentType: contentType}
	s.mu.Unlock()
	return nil
}

func (s *FakeStorage) Get(_ context.Context, key string) (io.ReadCloser, media.ObjectInfo, error) {
	s.mu.RLock()
	obj, ok := s.objects[key]
	s.mu.RUnlock()
	if !ok {
		return nil, media.ObjectInfo{}, media.ErrObjectNotFound
	}
	return io.NopCloser(bytes.NewReader(obj.data)), media.ObjectInfo{
		ContentType: obj.contentType,
		Size:        int64(len(obj.data)),
	}, nil
}

// Has reports whether an object with the given key was stored.
func (s *FakeStorage) Has(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.objects[key]
	return ok
}

// Keys returns all stored object keys (order is unspecified).
func (s *FakeStorage) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]string, 0, len(s.objects))
	for k := range s.objects {
		keys = append(keys, k)
	}
	return keys
}
