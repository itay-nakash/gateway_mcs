package multicluster_gw

import (
	"sync"

	"k8s.io/apimachinery/pkg/api/errors"
)

// Basic set implementasion, used as the dataset for saving the existing ServiceImports

type void struct{}

var member void

type Set struct {
	Elements map[string]void
	mutex    *sync.RWMutex
}

func NewSiSet() *Set {
	var set Set
	set.Elements = make(map[string]void)
	set.mutex = new(sync.RWMutex)
	return &set
}

func (s *Set) Add(elem string) {
	// write - so I use 'regular' lock
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.Elements[elem] = member
}

func (s *Set) Delete(elem string) error {
	// write - so I use 'regular' lock
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if _, exists := s.Elements[elem]; !exists {
		// #TODO check about the error:
		return errors.NewBadRequest("Service Import is not present in set")
	}
	delete(s.Elements, elem)
	return nil
}

// can called by multicluster_gw (the plugin) when checking if a spesific SI is exsits
// Therefore, needed to be sync:
func (s *Set) Contains(elem string) bool {
	// read - so I use RLock
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	_, exists := s.Elements[elem]
	return exists
}

func (s *Set) GetSize() int {
	// read - so I use RLock
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.Elements)
}
